package standard

type MetadataSchema struct {
	Name                 string `json:"name"`
	SpecificationVersion string `json:"specificationVersion"`
}

type V1SpecificationSchema struct {
}

type Schema struct {
	Metadata      MetadataSchema `json:"metadata"`
	Specification any            `json:"specification"`
}
