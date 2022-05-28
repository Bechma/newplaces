package backend

type Pixel struct {
	X     uint32 `key:"required" json:"x"`
	Y     uint32 `key:"required" json:"y"`
	Color uint32 `key:"required" json:"color"`
}
