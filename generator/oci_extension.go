package generator

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/field"
	maputils "github.com/ARM-software/golang-utils/utils/maps"
	"github.com/ARM-software/golang-utils/utils/reflection"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
)

type OciMetadata struct {
	Name  string
	Value string
}

type OCIImageDistributionInfo struct {
	PublishesToGHCR bool
	ImageName       string
	PackagePageURL  string
	ReleaseBadgeURL string
	SourceURL       string
	Title           string
	Description     string
	Documentation   string
	Licences        string
	Labels          []OciMetadata
	Attributes      []OciMetadata
}

func precomputeOCIImageDistribution(spec *SDKSpecification) OCIImageDistributionInfo {
	labels := ociLabels(spec)
	attributes := ociAnnotations(spec)

	return OCIImageDistributionInfo{
		PublishesToGHCR: publishesOCIToGHCR(spec),
		ImageName:       ghcrImageName(spec),
		PackagePageURL:  ghcrPackagePageURL(spec),
		ReleaseBadgeURL: ghcrReleaseBadgeURL(spec),
		SourceURL:       ociMetadataValue(labels, ociv1.AnnotationSource),
		Title:           ociMetadataValue(labels, ociv1.AnnotationTitle),
		Description:     ociMetadataValue(labels, ociv1.AnnotationDescription),
		Documentation:   ociMetadataValue(labels, ociv1.AnnotationDocumentation),
		Licences:        ociMetadataValue(labels, ociv1.AnnotationLicenses),
		Labels:          labels,
		Attributes:      attributes,
	}
}

type DistributionInfo struct {
	ToolSkillsPath string
}

func precomputeDistribution(spec *SDKSpecification) DistributionInfo {
	return DistributionInfo{
		ToolSkillsPath: toolSkillsInstallPath(spec),
	}
}

func hasOCIDistribution(spec *SDKSpecification) bool {
	return spec != nil && spec.DistributeOci != nil
}

func publishesOCIToGHCR(spec *SDKSpecification) bool {
	return hasOCIDistribution(spec) && spec.CommandLineCommands != nil && isHostedOnGitHub(spec)
}

func ghcrImageName(spec *SDKSpecification) string {
	if !publishesOCIToGHCR(spec) {
		return ""
	}

	repository := strings.ToLower(strings.TrimSpace(repositoryLocationOnGitHub(spec)))
	if reflection.IsEmpty(repository) {
		return ""
	}

	return fmt.Sprintf("ghcr.io/%s", repository)
}

func ghcrReleaseBadgeURL(spec *SDKSpecification) string {
	if !publishesOCIToGHCR(spec) {
		return ""
	}

	return fmt.Sprintf("https://img.shields.io/github/v/release/%s?label=container", repositoryLocationOnGitHub(spec))
}

func ghcrPackagePageURL(spec *SDKSpecification) string {
	if !publishesOCIToGHCR(spec) {
		return ""
	}

	owner := repositoryOrganisationOnGitHub(spec)
	packageName := repositoryNameOnGitHub(spec)
	if reflection.IsEmpty(owner) || reflection.IsEmpty(packageName) {
		return ""
	}

	return fmt.Sprintf("https://github.com/orgs/%s/packages/container/%s", owner, packageName)
}

func toolSkillsInstallPath(spec *SDKSpecification) string {
	if spec == nil || spec.CommandLineCommands == nil {
		return ""
	}

	binary := strings.TrimSpace(field.OptionalString(spec.CommandLineCommands.Binary, ""))
	if reflection.IsEmpty(binary) {
		return ""
	}

	return fmt.Sprintf("/etc/%s/.agents/skills", binary)
}

func ociLabels(spec *SDKSpecification) []OciMetadata {
	if spec == nil {
		return nil
	}

	title := strings.TrimSpace(field.OptionalString(spec.Name, spec.PackageName))
	description := strings.TrimSpace(field.OptionalString(spec.Description, ""))
	if spec.DistributeOci != nil && reflection.IsNotEmpty(spec.DistributeOci.Description) {
		description = strings.TrimSpace(spec.DistributeOci.Description)
	}
	documentation := repositoryURL(spec)

	defaults := map[string]string{
		ociv1.AnnotationDescription:   description,
		ociv1.AnnotationDocumentation: documentation,
		ociv1.AnnotationLicenses:      strings.TrimSpace(spec.Licence),
		ociv1.AnnotationSource:        repositoryURL(spec),
		ociv1.AnnotationTitle:         title,
	}

	if spec.DistributeOci == nil {
		return mergeOCIMetadata(defaults, nil)
	}

	return mergeOCIMetadata(defaults, stringMapFromOCIMap(spec.DistributeOci.Labels))
}

func ociAnnotations(spec *SDKSpecification) []OciMetadata {
	if spec == nil || spec.DistributeOci == nil {
		return nil
	}

	title := strings.TrimSpace(field.OptionalString(spec.Name, spec.PackageName))
	documentation := repositoryURL(spec)

	defaults := map[string]string{
		ociv1.AnnotationCreated:         "{{ .Date }}",
		ociv1.AnnotationDescription:     strings.TrimSpace(spec.DistributeOci.Description),
		ociv1.AnnotationDocumentation:   documentation,
		ociv1.AnnotationLicenses:        strings.TrimSpace(spec.Licence),
		"org.opencontainers.image.name": "{{ .ProjectName }}",
		ociv1.AnnotationRevision:        "{{ .FullCommit }}",
		ociv1.AnnotationSource:          repositoryURL(spec),
		ociv1.AnnotationTitle:           title,
		ociv1.AnnotationVersion:         "{{ .Version }}",
	}

	return mergeOCIMetadata(defaults, stringMapFromOCIMap(spec.DistributeOci.Attributes))
}

func mergeOCIMetadata(defaults map[string]string, overrides map[string]string) []OciMetadata {
	merged := maputils.Merge(defaults, overrides)

	entries := collection.Map(slices.Collect(maps.Keys(merged)), func(key string) OciMetadata {
		return OciMetadata{Name: key, Value: merged[key]}
	})
	slices.SortFunc(entries, func(left, right OciMetadata) int {
		return strings.Compare(left.Name, right.Name)
	})

	return entries
}

func ociMetadataValue(entries []OciMetadata, name string) string {
	if entry, found := collection.FirstBy(entries, func(entry OciMetadata) bool {
		return entry.Name == name
	}); found {
		return entry.Value
	}
	return ""
}

func stringMapFromOCIMap(entries OciMapStringString) map[string]string {
	if len(entries) == 0 {
		return nil
	}

	return collection.Reduce(slices.Collect(collection.FilterSequence(maps.Keys(entries), func(key string) bool {
		value, ok := entries[key]
		if !ok {
			return false
		}
		_, err := validation.EnsureString(value)
		return err == nil && reflection.IsNotEmpty(key)
	})), map[string]string{}, func(acc map[string]string, key string) map[string]string {
		stringValue, _ := validation.EnsureString(entries[key])
		acc[strings.TrimSpace(key)] = strings.TrimSpace(stringValue)
		return acc
	})
}
