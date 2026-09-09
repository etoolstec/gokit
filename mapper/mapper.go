package mapper

import "github.com/jinzhu/copier"

func MapToModel(src any, dst any) error {
	return copier.CopyWithOption(dst, src, copier.Option{IgnoreEmpty: true, DeepCopy: true})
}
func MapToDTO(src any, dst any) error {
	return copier.CopyWithOption(dst, src, copier.Option{IgnoreEmpty: true, DeepCopy: true})
}
