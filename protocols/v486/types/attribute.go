package types

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// Attribute is an entity attribute, that holds specific data such as the health of the entity. Each attribute
// holds a default value, maximum and minimum value, name and its current value.
type Attribute struct {
	protocol.Attribute
}

// Marshal encodes/decodes an Attribute.
func (x *Attribute) Marshal(r protocol.IO) {
	r.Float32(&x.Min)
	r.Float32(&x.Max)
	r.Float32(&x.Value)
	r.Float32(&x.Default)
	r.String(&x.Name)
}
