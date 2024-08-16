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

// GetBuilder 返回一个 GetOption 类型的函数，用于设置 GetOptions 的 Builders 字段
func GetBuilder(b *Builder) GetOption {
	return func(options *GetOptions) {
		// 如果 LinkType 未设置，则默认设置为 LinkTypeAND
		if b.LinkType == LinkTypeNotSet {
			b.LinkType = LinkTypeAND
		}

		// 检查 LinkType 是否为支持的类型，如果不是则抛出错误
		if b.LinkType != LinkTypeAND && b.LinkType != LinkTypeOR {
			panic(fmt.Errorf("not supported link type %s", b.LinkType))
		}

		// 将 Builder 添加到 GetOptions 的 Builders 列表中
		options.Builders = append(options.Builders, b)
	}
}
