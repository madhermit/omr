// Package plan computes and executes the symlink/restart plan for a worktree switch.
package plan

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/madhermit/omr/internal/config"
	"github.com/madhermit/omr/internal/git"
	"github.com/madhermit/omr/internal/overmind"
	"github.com/madhermit/omr/internal/probe"
	"github.com/madhermit/omr/internal/symlink"
)

// DirFlip is one symlink update: change the symlink at cfg.Root/Dir from OldTarget to NewTarget.
// OldTarget is empty if the symlink was missing or broken.
type DirFlip struct {
	Dir       string
	OldTarget string
	NewTarget string
}

// Plan describes what a switch will do: symlink flips and ordered restart waves.
type Plan struct {
	DirFlips []DirFlip
	Waves    [][]string // service names per wave (wave 0 starts first)
}

func (p Plan) IsEmpty() bool { return len(p.DirFlips) == 0 && len(p.Waves) == 0 }

func (p Plan) AnyRestarts() bool {
	for _, w := range p.Waves {
		if len(w) > 0 {
			return true
		}
	}
	return false
}

type Logger struct {
	Logf  func(format string, args ...interface{})
	Warnf func(format string, args ...interface{})
}

func (l Logger) logf(format string, args ...interface{}) {
	if l.Logf != nil {
		l.Logf(format, args...)
	}
}

func (l Logger) warnf(format string, args ...interface{}) {
	if l.Warnf != nil {
		l.Warnf(format, args...)
	}
}

// Build computes the symlink flips and ordered restart waves for a switch.
// Resolver returns each service's (oldTarget, newTarget) — production callers
// use BuildResolver; tests inject their own.
func Build(cfg *config.Config, services []string, resolver Resolver) (Plan, error) {
	resolved, err := resolveAll(cfg, services, resolver)
	if err != nil {
		return Plan{}, err
	}

	flips := computeFlips(resolved)
	included := selectServices(resolved)
	waves := BuildWaves(cfg, included)

	return Plan{DirFlips: flips, Waves: waves}, nil
}

// Resolver returns each service's current symlink target and new worktree path.
// oldTarget is empty if the symlink is missing or broken.
type Resolver func(svc string) (oldTarget, newTarget string, err error)

// BuildResolver wires plan to symlink+git. Worktree lists are cached per search
// directory so a monorepo with N services runs `git worktree list` once.
func BuildResolver(cfg *config.Config, branch string) Resolver {
	worktreeCache := map[string][]git.Worktree{}
	return func(svcName string) (string, string, error) {
		svc, ok := cfg.Services[svcName]
		if !ok {
			return "", "", fmt.Errorf("unknown service: %s", svcName)
		}

		linkPath := filepath.Join(cfg.Root, svc.Dir)
		valid, currentTarget, _ := symlink.Verify(linkPath)
		oldTarget := ""
		searchDir := filepath.Dir(linkPath)
		if valid && currentTarget != "" {
			oldTarget = currentTarget
			searchDir = currentTarget
		}

		worktrees, ok := worktreeCache[searchDir]
		if !ok {
			var err error
			worktrees, err = git.ListWorktreesInDir(searchDir)
			if err != nil {
				return "", "", fmt.Errorf("listing worktrees for %s: %w", svcName, err)
			}
			worktreeCache[searchDir] = worktrees
		}
		for _, wt := range worktrees {
			if wt.Branch == branch {
				return oldTarget, wt.Path, nil
			}
		}
		return "", "", fmt.Errorf("no worktree for %s branch %q", svcName, branch)
	}
}

type resolvedService struct {
	name      string
	dir       string
	oldTarget string
	newTarget string
}

func resolveAll(cfg *config.Config, services []string, resolver Resolver) ([]resolvedService, error) {
	out := make([]resolvedService, 0, len(services))
	for _, name := range services {
		svc := cfg.Services[name]
		oldT, newT, err := resolver(name)
		if err != nil {
			return nil, err
		}
		out = append(out, resolvedService{
			name:      name,
			dir:       svc.Dir,
			oldTarget: oldT,
			newTarget: newT,
		})
	}
	return out, nil
}

func computeFlips(resolved []resolvedService) []DirFlip {
	seen := map[string]bool{}
	var flips []DirFlip
	for _, r := range resolved {
		if seen[r.dir] {
			continue
		}
		seen[r.dir] = true
		if r.oldTarget == r.newTarget {
			continue
		}
		flips = append(flips, DirFlip{Dir: r.dir, OldTarget: r.oldTarget, NewTarget: r.newTarget})
	}
	return flips
}

// selectServices includes any service whose symlink target is changing (or
// unknown). Services whose symlink already points at the new worktree are
// presumed correctly running and skipped.
func selectServices(resolved []resolvedService) []string {
	var included []string
	for _, r := range resolved {
		if r.oldTarget == "" || r.oldTarget != r.newTarget {
			included = append(included, r.name)
		}
	}
	return included
}

// BuildWaves layers included services by depends_on (Kahn). Edges pointing
// outside the included set are dropped — those services are presumed running.
func BuildWaves(cfg *config.Config, included []string) [][]string {
	if len(included) == 0 {
		return nil
	}

	includedSet := map[string]bool{}
	for _, n := range included {
		includedSet[n] = true
	}

	remainingDeps := map[string]map[string]bool{}
	dependents := map[string][]string{}
	for _, name := range included {
		svc := cfg.Services[name]
		deps := map[string]bool{}
		for _, d := range svc.DependsOn {
			if includedSet[d] {
				deps[d] = true
				dependents[d] = append(dependents[d], name)
			}
		}
		remainingDeps[name] = deps
	}

	var waves [][]string
	remaining := map[string]bool{}
	for _, n := range included {
		remaining[n] = true
	}

	for len(remaining) > 0 {
		var wave []string
		for name := range remaining {
			if len(remainingDeps[name]) == 0 {
				wave = append(wave, name)
			}
		}
		if len(wave) == 0 {
			// Cycle — should have been caught by config validation, but be defensive.
			for name := range remaining {
				wave = append(wave, name)
			}
			sort.Strings(wave)
			waves = append(waves, wave)
			return waves
		}
		sort.Strings(wave)
		for _, name := range wave {
			delete(remaining, name)
			for _, dependent := range dependents[name] {
				delete(remainingDeps[dependent], name)
			}
		}
		waves = append(waves, wave)
	}
	return waves
}

// Execute flips symlinks, then restarts each wave and waits for it to be ready.
// Each wave probes under its own waveTimeout — slow earlier waves don't starve
// later ones. Aborts on the first wave's probe failure.
func Execute(ctx context.Context, cfg *config.Config, p Plan, waveTimeout time.Duration, log Logger) error {
	for _, flip := range p.DirFlips {
		if err := symlink.Create(cfg.Root, flip.Dir, flip.NewTarget); err != nil {
			return fmt.Errorf("creating symlink for %s: %w", flip.Dir, err)
		}
	}

	if !p.AnyRestarts() {
		return nil
	}

	if !overmind.IsRunning(cfg.Root) {
		log.warnf("Overmind is not running. Start it with: overmind start")
		return nil
	}

	for i, wave := range p.Waves {
		procs := collectProcs(cfg, wave)
		if len(procs) == 0 {
			continue
		}
		log.logf("Wave %d: restarting %v\n", i+1, wave)
		if err := overmind.Restart(cfg.Root, procs...); err != nil {
			return fmt.Errorf("wave %d: %w", i+1, err)
		}
		waveCtx, cancel := context.WithTimeout(ctx, waveTimeout)
		err := waitWaveReady(waveCtx, cfg, wave, log)
		cancel()
		if err != nil {
			return fmt.Errorf("wave %d not ready: %w", i+1, err)
		}
	}
	return nil
}

func collectProcs(cfg *config.Config, services []string) []string {
	var procs []string
	for _, name := range services {
		procs = append(procs, cfg.Services[name].Procs...)
	}
	return procs
}

// waitWaveReady probes every service in the wave with a configured port in parallel.
func waitWaveReady(ctx context.Context, cfg *config.Config, wave []string, log Logger) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(wave))

	for _, name := range wave {
		svc := cfg.Services[name]
		if svc.Port == 0 {
			continue
		}
		wg.Add(1)
		go func(n string, port int) {
			defer wg.Done()
			log.logf("  waiting for %s on :%d…\n", n, port)
			if err := probe.WaitReady(ctx, port); err != nil {
				errCh <- fmt.Errorf("%s (:%d): %w", n, port, err)
				return
			}
			log.logf("  %s ready\n", n)
		}(name, svc.Port)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}
