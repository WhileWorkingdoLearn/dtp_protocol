package dtp

import (
	"fmt"
)

type PacketError struct {
	Text     string
	Want     int
	Has      int
	PacketID int
}

func (pe PacketError) Error() string {
	return fmt.Sprintf("error: %s. got: %d, want: %d, package:%v \n", pe.Text, pe.Has, pe.Want, pe.PacketID)
}
