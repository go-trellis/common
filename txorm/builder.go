package txorm

import "fmt"

type LinkType string

const (
	LinkTypeNotSet LinkType = ""
	LinkTypeAND    LinkType = "AND"
	LinkTypeOR     LinkType = "OR"
)

type Builder struct {
	LinkType LinkType
	Where    interface{}
	Args     []interface{}
}

func GetBuilder(b *Builder) GetOption {
	return func(options *GetOptions) {
		if b.LinkType == LinkTypeNotSet {
			b.LinkType = LinkTypeAND
		}
		if b.LinkType != LinkTypeAND && b.LinkType != LinkTypeOR {
			panic(fmt.Errorf("not supported link type %s", b.LinkType))
		}
		options.Builders = append(options.Builders, b)
	}
}
