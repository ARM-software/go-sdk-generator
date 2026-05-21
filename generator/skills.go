package generator

import (
	"context"
	"embed"
	"maps"
	"slices"
	"strings"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

const embeddedSkillsDir = "skills"

var embeddedAgentSkillPaths = []string{
	"run-go-sdk-generator/SKILL.md",
	"using-go-sdk-generator-in-a-github-action/SKILL.md",
	"install-go-sdk-generator/SKILL.md",
}

var (
	//go:embed skills/*/SKILL.md
	embeddedSkillsFS embed.FS
)

func EmbeddedAgentSkills() (skills map[string][]byte, err error) {
	skills = make(map[string][]byte, len(embeddedAgentSkillPaths))

	embeddedFs, err := filesystem.NewEmbedFileSystem(&embeddedSkillsFS)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not initialise embedded skills filesystem")
		return
	}

	for _, skillPath := range embeddedAgentSkillPaths {
		embeddedPath := filesystem.FilePathJoin(embeddedFs, embeddedSkillsDir, skillPath)

		skill, subErr := embeddedFs.ReadFile(embeddedPath)
		if subErr != nil {
			err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, subErr, "could not read embedded skill %q", embeddedPath)
			return
		}

		skills[skillPath] = skill
	}

	return
}

func InitialiseAgentSkills(ctx context.Context, outputDir string) (err error) {
	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	if outputDir == "" {
		err = commonerrors.UndefinedVariable("output directory")
		return
	}

	localFs := filesystem.GetGlobalFileSystem()

	skills, err := EmbeddedAgentSkills()
	if err != nil {
		return
	}

	agentsDir := filesystem.FilePathJoin(localFs, outputDir, ".agents")
	if !filesystem.Exists(agentsDir) {
		err = filesystem.MkDir(agentsDir)
		if err != nil {
			err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not create %q", agentsDir)
			return
		}
	}

	skillsDir := filesystem.FilePathJoin(localFs, agentsDir, "skills")
	if !filesystem.Exists(skillsDir) {
		err = filesystem.MkDir(skillsDir)
		if err != nil {
			err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not create %q", skillsDir)
			return
		}
	}

	for _, skillPath := range slices.Sorted(maps.Keys(skills)) {
		if err = parallelisation.DetermineContextError(ctx); err != nil {
			return
		}

		skillFilePath := filesystem.FilePathJoin(localFs, append([]string{skillsDir}, strings.Split(skillPath, "/")...)...)
		skillDir := filesystem.FilePathDir(localFs, skillFilePath)
		if !filesystem.Exists(skillDir) {
			err = filesystem.MkDir(skillDir)
			if err != nil {
				err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not create %q", skillDir)
				return
			}
		}

		err = filesystem.WriteFile(skillFilePath, skills[skillPath], 0644)
		if err != nil {
			err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not write skill %q", skillPath)
			return
		}
	}

	return
}

func InitialiseAgentSkillsInCurrentDirectory(ctx context.Context) (err error) {
	if err = parallelisation.DetermineContextError(ctx); err != nil {
		return
	}

	workingDir, err := filesystem.CurrentDirectory()
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnexpected, err, "could not determine the current directory")
		return
	}

	err = InitialiseAgentSkills(ctx, workingDir)
	return
}
