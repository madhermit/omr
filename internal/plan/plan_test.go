package plan

import (
	"reflect"
	"testing"

	"github.com/madhermit/omr/internal/config"
)

func staticResolver(m map[string][2]string) Resolver {
	return func(svc string) (string, string, error) {
		pair, ok := m[svc]
		if !ok {
			return "", "", nil
		}
		return pair[0], pair[1], nil
	}
}

func TestBuild_MonorepoOneFlipManyRestarts(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.Service{
			"api":     {Dir: "current", Procs: []string{"rails"}, Port: 9002},
			"web":     {Dir: "current", Procs: []string{"web"}, Port: 3002, DependsOn: []string{"api"}},
			"landing": {Dir: "current", Procs: []string{"landing"}},
		},
	}
	r := staticResolver(map[string][2]string{
		"api":     {"/main", "/feature"},
		"web":     {"/main", "/feature"},
		"landing": {"/main", "/feature"},
	})

	plan, err := Build(cfg, []string{"api", "web", "landing"}, r)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if len(plan.DirFlips) != 1 {
		t.Errorf("expected 1 flip (shared dir), got %d", len(plan.DirFlips))
	}
	if got := plan.DirFlips[0].Dir; got != "current" {
		t.Errorf("flip Dir: %q vs %q", got, "current")
	}
	if len(plan.Waves) != 2 {
		t.Fatalf("expected 2 waves (api before web; landing parallel with api), got %v", plan.Waves)
	}
	if !reflect.DeepEqual(plan.Waves[0], []string{"api", "landing"}) {
		t.Errorf("wave 0 = %v, want [api landing]", plan.Waves[0])
	}
	if !reflect.DeepEqual(plan.Waves[1], []string{"web"}) {
		t.Errorf("wave 1 = %v, want [web]", plan.Waves[1])
	}
}

func TestBuild_SameTarget_NoFlipNoRestart(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.Service{
			"api": {Dir: "current"},
		},
	}
	r := staticResolver(map[string][2]string{
		"api": {"/some/path", "/some/path"},
	})

	plan, err := Build(cfg, []string{"api"}, r)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.IsEmpty() {
		t.Errorf("expected empty plan, got: %+v", plan)
	}
}

func TestBuild_BrokenSymlink_IncludesAll(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.Service{
			"api": {Dir: "current"},
		},
	}
	r := staticResolver(map[string][2]string{
		"api": {"", "/feature"},
	})

	plan, err := Build(cfg, []string{"api"}, r)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Waves) != 1 || !reflect.DeepEqual(plan.Waves[0], []string{"api"}) {
		t.Errorf("expected [api] included (no old target), got %v", plan.Waves)
	}
}

func TestBuildWaves_ThreeLevelChain(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.Service{
			"db":  {Dir: "a"},
			"api": {Dir: "a", DependsOn: []string{"db"}},
			"web": {Dir: "a", DependsOn: []string{"api"}},
		},
	}
	waves := BuildWaves(cfg, []string{"db", "api", "web"})
	if len(waves) != 3 {
		t.Fatalf("expected 3 waves, got %d: %v", len(waves), waves)
	}
	expected := [][]string{{"db"}, {"api"}, {"web"}}
	for i, w := range expected {
		if !reflect.DeepEqual(waves[i], w) {
			t.Errorf("wave %d: got %v, want %v", i, waves[i], w)
		}
	}
}

func TestBuildWaves_DepOutsideIncludedSet_Dropped(t *testing.T) {
	cfg := &config.Config{
		Services: map[string]config.Service{
			"api": {Dir: "a"},
			"web": {Dir: "a", DependsOn: []string{"api"}},
		},
	}
	// api isn't in the included set → web's dep is dropped → web in wave 0.
	waves := BuildWaves(cfg, []string{"web"})
	if len(waves) != 1 || !reflect.DeepEqual(waves[0], []string{"web"}) {
		t.Errorf("expected [[web]], got %v", waves)
	}
}
