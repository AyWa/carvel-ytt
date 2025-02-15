// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package filepos

import (
	"fmt"
	"strings"
)

type Position struct {
	lineNum    *int // 1 based
	file       string
	line       string
	known      bool
	fromMemory bool
}

func NewPosition(lineNum int) *Position {
	if lineNum <= 0 {
		panic("Lines are 1 based")
	}
	return &Position{lineNum: &lineNum, known: true}
}

// NewUnknownPosition is equivalent of zero value *Position
func NewUnknownPosition() *Position {
	return &Position{}
}

func NewUnknownPositionWithKeyVal(k, v interface{}, separator string) *Position {
	return &Position{line: fmt.Sprintf("%v%v %#v", k, separator, v), fromMemory: true}
}

func (p *Position) SetFile(file string) { p.file = file }

func (p *Position) SetLine(line string) { p.line = line }

func (p *Position) IsKnown() bool { return p != nil && p.known }

func (p *Position) FromMemory() bool { return p.fromMemory }

func (p *Position) LineNum() int {
	if !p.IsKnown() {
		panic("Position is unknown")
	}
	if p.lineNum == nil {
		panic("Position was not properly initialized")
	}
	return *p.lineNum
}

func (p *Position) GetLine() string {
	return p.line
}

func (p *Position) AsString() string {
	return "line " + p.AsCompactString()
}

func (p *Position) GetFile() string {
	return p.file
}

func (p *Position) AsCompactString() string {
	filePrefix := p.file
	if len(filePrefix) > 0 {
		filePrefix += ":"
	}
	if p.IsKnown() {
		return fmt.Sprintf("%s%d", filePrefix, p.LineNum())
	}
	return fmt.Sprintf("%s?", filePrefix)
}

func (p *Position) AsIntString() string {
	if p.IsKnown() {
		return fmt.Sprintf("%d", p.LineNum())
	}
	return "?"
}

func (p *Position) As4DigitString() string {
	if p.IsKnown() {
		return fmt.Sprintf("%4d", p.LineNum())
	}
	return "????"
}

func (p *Position) DeepCopy() *Position {
	if p == nil {
		return nil
	}
	newPos := &Position{file: p.file, known: p.known, line: p.line}
	if p.lineNum != nil {
		lineVal := *p.lineNum
		newPos.lineNum = &lineVal
	}
	return newPos
}

func (p *Position) DeepCopyWithLineOffset(offset int) *Position {
	if !p.IsKnown() {
		panic("Position is unknown")
	}
	if offset < 0 {
		panic("Unexpected line offset")
	}
	newPos := p.DeepCopy()
	*newPos.lineNum += offset
	return newPos
}

// IsNextTo compares the location of one position with another.
func (p *Position) IsNextTo(otherPostion *Position) bool {
	if p.IsKnown() && otherPostion.IsKnown() {
		if p.GetFile() == otherPostion.GetFile() {
			diff := p.LineNum() - otherPostion.LineNum()
			if -1 <= diff && 1 >= diff {
				return true
			}
		}
	}
	return false
}

// PopulateAnnotationPositionFromNode takes an annotation's position, the position of the node it annotates,
// and the comments from the same node, and creates an approximation of the annotation file and line string.
func PopulateAnnotationPositionFromNode(annPos *Position, nodePosition *Position, nodeComments []Meta) *Position {
	leftPadding := 0
	if nodePosition.IsKnown() {
		nodeLine := nodePosition.GetLine()
		leftPadding = len(nodeLine) - len(strings.TrimLeft(nodeLine, " "))
	}

	lineString := ""
	for _, c := range nodeComments {
		if c.Position.IsKnown() && c.Position.AsIntString() == fmt.Sprintf("%d", annPos.LineNum()) {
			lineString = fmt.Sprintf("%v#%s", strings.Repeat(" ", leftPadding), c.Data)
		}
	}

	annPos.SetFile(nodePosition.GetFile())
	annPos.SetLine(lineString)

	return annPos
}

// Meta is a representation of comment's source, contains comment as a string, and original position.
// Used to populate
type Meta struct {
	Data     string
	Position *Position
}
