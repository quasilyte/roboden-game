package assets

import (
	"runtime"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
)

func RegisterShaderResources(ctx *ge.Context, config *Config, progress *float64) {
	// Associate shader resources.
	shaderResources := map[resource.ShaderID]resource.ShaderInfo{
		ShaderDissolve:         {Path: "shader/dissolve.go"},
		ShaderLavaPuddle:       {Path: "shader/lava_puddle.go"},
		ShaderColonyBuild:      {Path: "shader/colony_build.go"},
		ShaderTurretBuild:      {Path: "shader/turret_build.go"},
		ShaderColonyDamage:     {Path: "shader/colony_damage.go"},
		ShaderCreepTurretBuild: {Path: "shader/creep_turret_build.go"},
		ShaderSlideX:           {Path: "shader/slide_x.go"},
		ShaderColonyTeleport:   {Path: "shader/colony_teleport.go"},
	}

	singleThread := runtime.GOMAXPROCS(-1) == 1
	progressPerItem := 1.0 / float64(len(shaderResources))
	for id, res := range shaderResources {
		ctx.Loader.ShaderRegistry.Set(id, res)
		ctx.Loader.LoadShader(id)
		if progress != nil {
			*progress += progressPerItem
		}
		if singleThread {
			runtime.Gosched()
		}
	}
}

const (
	ShaderNone resource.ShaderID = iota
	ShaderDissolve
	ShaderLavaPuddle
	ShaderColonyBuild
	ShaderTurretBuild
	ShaderColonyDamage
	ShaderCreepTurretBuild
	ShaderSlideX
	ShaderColonyTeleport
)
