package matchArxiv

type refMatch struct {
	ArxivID string `bson:"arxivID,omitempty"`
	Mode    int32  `bson:"mode"`
}

type wikipediaRefObj struct {
	ID          string            `bson:"_id"`
	ArtIDSnap   []int64           `bson:"artIDSnap"`
	ArtDateSnap interface{}       `bson:"artDateSnap"`
	DateSnap    []int32           `bson:"dateSnap"`
	Ref         map[string]string `bson:"ref"`
	Match       refMatch          `bson:"match,omitempty"`
}
