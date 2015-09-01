// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package recorder provides support for vector graphics serialization.
package recorder

import (
	"fmt"
	"image/color"
	"runtime"

	"github.com/sksullivan/plot/vg"
)

var _ vg.Canvas = (*Canvas)(nil)

// Canvas implements vg.Canvas operation serialization.
type Canvas struct {
	// Resolution holds the canvas resolution in DPI.
	Resolution float64

	// Actions holds a log of all methods called on
	// the canvas.
	Actions []Action

	// KeepCaller indicates whether the Canvas will
	// retain runtime caller location for the actions.
	// This includes source filename and line number.
	KeepCaller bool

	// c holds a backing vg.Canvas. If c is non-nil
	// then method calls to the Canvas will be
	// reflected to c.
	c vg.Canvas

	// fonts holds a collection of font/size descriptions.
	fonts map[fontID]vg.Font
}

type fontID struct {
	name string
	size vg.Length
}

// Action is a vector graphics action as defined by the
// vg.Canvas interface. Each method of Canvas has a
// corresponding Action type.
type Action interface {
	Call() string
	ApplyTo(vg.Canvas)
	callerLocation() *callerLocation
}

type callerLocation struct {
	haveCaller bool
	file       string
	line       int
}

func (l *callerLocation) set() {
	_, l.file, l.line, l.haveCaller = runtime.Caller(3)
}

func (l callerLocation) String() string {
	if !l.haveCaller {
		return ""
	}
	return fmt.Sprintf("%s:%d ", l.file, l.line)
}

// New returns a new Canvas with the specified resolution.
func New(dpi float64) *Canvas { return &Canvas{Resolution: dpi} }

// NewFrom returns a new Canvas from an existing vg.Canvas. vg.Canvas methods
// called on the Canvas will also be passed to the provided backing
// vg.Canvas. The Resolution field is set to the value returned by c.DPI.
func NewFrom(c vg.Canvas) *Canvas { return &Canvas{Resolution: c.DPI(), c: c} }

// Reset resets the Canvas to the base state.
func (c *Canvas) Reset() {
	c.Actions = c.Actions[:0]
}

// ReplayOn applies the set of Actions recorded by the Recorder onto
// the destination Canvas.
func (c *Canvas) ReplayOn(dst vg.Canvas) error {
	if c.fonts == nil {
		c.fonts = make(map[fontID]vg.Font)
	}
	for _, a := range c.Actions {
		fa, ok := a.(*FillString)
		if !ok {
			continue
		}
		f := fontID{name: fa.Font, size: fa.Size}
		if _, exists := c.fonts[f]; !exists {
			var err error
			c.fonts[f], err = vg.MakeFont(fa.Font, fa.Size)
			if err != nil {
				return err
			}
		}
		fa.fonts = c.fonts
	}
	for _, a := range c.Actions {
		a.ApplyTo(dst)
	}
	return nil
}

func (c *Canvas) append(a Action) {
	if c.c != nil {
		a.ApplyTo(c)
	}
	if c.KeepCaller {
		a.callerLocation().set()
	}
	c.Actions = append(c.Actions, a)
}

// SetLineWidth corresponds to the vg.Canvas.SetWidth method.
type SetLineWidth struct {
	Width vg.Length

	l callerLocation
}

// SetLineWidth implements the SetLineWidth method of the vg.Canvas interface.
func (c *Canvas) SetLineWidth(w vg.Length) {
	c.append(&SetLineWidth{Width: w})
}

// Call returns the method call that generated the action.
func (a *SetLineWidth) Call() string {
	return fmt.Sprintf("%sSetLineWidth(%v)", a.l, a.Width)
}

// ApplyTo applies the action to the given vg.Canvas.
func (a *SetLineWidth) ApplyTo(c vg.Canvas) {
	c.SetLineWidth(a.Width)
}

func (a *SetLineWidth) callerLocation() *callerLocation {
	return &a.l
}

// SetLineDash corresponds to the vg.Canvas.SetLineDash method.
type SetLineDash struct {
	Dashes  []vg.Length
	Offsets vg.Length

	l callerLocation
}

// SetLineDash implements the SetLineDash method of the vg.Canvas interface.
func (c *Canvas) SetLineDash(dashes []vg.Length, offs vg.Length) {
	c.append(&SetLineDash{
		Dashes:  append([]vg.Length(nil), dashes...),
		Offsets: offs,
	})
}

// Call returns the method call that generated the action.
func (a *SetLineDash) Call() string {
	return fmt.Sprintf("%sSetLineDash(%#v, %v)", a.l, a.Dashes, a.Offsets)
}

// ApplyTo applies the action to the given vg.Canvas.
func (a *SetLineDash) ApplyTo(c vg.Canvas) {
	c.SetLineDash(a.Dashes, a.Offsets)
}

func (a *SetLineDash) callerLocation() *callerLocation {
	return &a.l
}

// SetColor corresponds to the vg.Canvas.SetColor method.
type SetColor struct {
	Color color.Color

	l callerLocation
}

// SetColor implements the SetColor method of the vg.Canvas interface.
func (c *Canvas) SetColor(col color.Color) {
	c.append(&SetColor{Color: col})
}

// Call returns the method call that generated the action.
func (a *SetColor) Call() string {
	return fmt.Sprintf("%sSetColor(%#v)", a.l, a.Color)
}

// ApplyTo applies the action to the given vg.Canvas.
func (a *SetColor) ApplyTo(c vg.Canvas) {
	c.SetColor(a.Color)
}

func (a *SetColor) callerLocation() *callerLocation {
	return &a.l
}

// Rotate corresponds to the vg.Canvas.Rotate method.
type Rotate struct {
	Angle float64

	l callerLocation
}

// Rotate implements the Rotate method of the vg.Canvas interface.
func (c *Canvas) Rotate(a float64) {
	c.append(&Rotate{Angle: a})
}

// Call returns the method call that generated the action.
func (a *Rotate) Call() string {
	return fmt.Sprintf("%sRotate(%v)", a.l, a.Angle)
}

// ApplyTo applies the action to the given vg.Canvas.
func (a *Rotate) ApplyTo(c vg.Canvas) {
	c.Rotate(a.Angle)
}

func (a *Rotate) callerLocation() *callerLocation {
	return &a.l
}

// Translate corresponds to the vg.Canvas.Translate method.
type Translate struct {
	X, Y vg.Length

	l callerLocation
}

// Translate implements the Translate method of the vg.Canvas interface.
func (c *Canvas) Translate(x, y vg.Length) {
	c.append(&Translate{X: x, Y: y})
}

// Call returns the method call that generated the action.
func (a *Translate) Call() string {
	return fmt.Sprintf("%sTranslate(%v, %v)", a.l, a.X, a.Y)
}

// ApplyTo applies the action to the given vg.Canvas.
func (a *Translate) ApplyTo(c vg.Canvas) {
	c.Translate(a.X, a.Y)
}

func (a *Translate) callerLocation() *callerLocation {
	return &a.l
}

// Scale corresponds to the vg.Canvas.Scale method.
type Scale struct {
	X, Y float64

	l callerLocation
}

// Scale implements the Scale method of the vg.Canvas interface.
func (c *Canvas) Scale(x, y float64) {
	c.append(&Scale{X: x, Y: y})
}

// Call returns the method call that generated the action.
func (a *Scale) Call() string {
	return fmt.Sprintf("%sScale(%v, %v)", a.l, a.X, a.Y)
}

// ApplyTo applies the action to the given vg.Canvas.
func (a *Scale) ApplyTo(c vg.Canvas) {
	c.Scale(a.X, a.Y)
}

func (a *Scale) callerLocation() *callerLocation {
	return &a.l
}

// Push corresponds to the vg.Canvas.Push method.
type Push struct {
	l callerLocation
}

// Push implements the Push method of the vg.Canvas interface.
func (c *Canvas) Push() {
	c.append(&Push{})
}

// Call returns the method call that generated the action.
func (a *Push) Call() string {
	return fmt.Sprintf("%sPush()", a.l)
}

// ApplyTo applies the action to the given vg.Canvas.
func (a *Push) ApplyTo(c vg.Canvas) {
	c.Push()
}

func (a *Push) callerLocation() *callerLocation {
	return &a.l
}

// Pop corresponds to the vg.Canvas.Pop method.
type Pop struct {
	l callerLocation
}

// Pop implements the Pop method of the vg.Canvas interface.
func (c *Canvas) Pop() {
	c.append(&Pop{})
}

// Call returns the method call that generated the action.
func (a *Pop) Call() string {
	return fmt.Sprintf("%sPop()", a.l)
}

// ApplyTo applies the action to the given vg.Canvas.
func (a *Pop) ApplyTo(c vg.Canvas) {
	c.Pop()
}

func (a *Pop) callerLocation() *callerLocation {
	return &a.l
}

// Stroke corresponds to the vg.Canvas.Stroke method.
type Stroke struct {
	Path vg.Path

	l callerLocation
}

// Stroke implements the Stroke method of the vg.Canvas interface.
func (c *Canvas) Stroke(path vg.Path) {
	c.append(&Stroke{Path: append(vg.Path(nil), path...)})
}

// Call returns the method call that generated the action.
func (a *Stroke) Call() string {
	return fmt.Sprintf("%sStroke(%#v)", a.l, a.Path)
}

// ApplyTo applies the action to the given vg.Canvas.
func (a *Stroke) ApplyTo(c vg.Canvas) {
	c.Stroke(a.Path)
}

func (a *Stroke) callerLocation() *callerLocation {
	return &a.l
}

// Fill corresponds to the vg.Canvas.Fill method.
type Fill struct {
	Path vg.Path

	l callerLocation
}

// Fill implements the Fill method of the vg.Canvas interface.
func (c *Canvas) Fill(path vg.Path) {
	c.append(&Fill{Path: append(vg.Path(nil), path...)})
}

// Call returns the method call that generated the action.
func (a *Fill) Call() string {
	return fmt.Sprintf("%sFill(%#v)", a.l, a.Path)
}

// ApplyTo applies the action to the given vg.Canvas.
func (a *Fill) ApplyTo(c vg.Canvas) {
	c.Fill(a.Path)
}

func (a *Fill) callerLocation() *callerLocation {
	return &a.l
}

// FillString corresponds to the vg.Canvas.FillString method.
type FillString struct {
	Font   string
	Size   vg.Length
	X, Y   vg.Length
	String string

	l callerLocation

	fonts map[fontID]vg.Font
}

// FillString implements the FillString method of the vg.Canvas interface.
func (c *Canvas) FillString(font vg.Font, x, y vg.Length, str string) {
	c.append(&FillString{
		Font: font.Name(),
		Size: font.Size,
		X:    x, Y: y,
		String: str,
	})
}

// ApplyTo applies the action to the given vg.Canvas.
func (a *FillString) ApplyTo(c vg.Canvas) {
	c.FillString(a.fonts[fontID{name: a.Font, size: a.Size}], a.X, a.Y, a.String)
}

// Call returns the pseudo method call that generated the action.
func (a *FillString) Call() string {
	return fmt.Sprintf("%sFillString(%q, %v, %v, %v, %q)", a.l, a.Font, a.Size, a.X, a.Y, a.String)
}

func (a *FillString) callerLocation() *callerLocation {
	return &a.l
}

// DPI corresponds to the vg.Canvas.DPI method.
type DPI struct {
	l callerLocation
}

// DPI implements the DPI method of the vg.Canvas interface.
func (c *Canvas) DPI() float64 {
	c.append(&DPI{})
	return c.Resolution
}

// Call returns the method call that generated the action.
func (a *DPI) Call() string {
	return fmt.Sprintf("%sDPI()", a.l)
}

// ApplyTo applies the action to the given vg.Canvas.
func (a *DPI) ApplyTo(c vg.Canvas) {
	c.DPI()
}

func (a *DPI) callerLocation() *callerLocation {
	return &a.l
}

// Commenter defines types that can record comments.
type Commenter interface {
	Comment(string)
}

var _ Commenter = (*Canvas)(nil)

// Comment implements a Recorder comment mechanism.
type Comment struct {
	Text string

	l callerLocation
}

// Comment adds a comment to a list of Actions..
func (c *Canvas) Comment(text string) {
	c.append(&Comment{Text: text})
}

// Call returns the method call that generated the action.
func (a *Comment) Call() string {
	return fmt.Sprintf("%sComment(%q)", a.l, a.Text)
}

// ApplyTo applies the action to the given vg.Canvas.
func (a *Comment) ApplyTo(c vg.Canvas) {
	if c, ok := c.(Commenter); ok {
		c.Comment(a.Text)
	}
}

func (a *Comment) callerLocation() *callerLocation {
	return &a.l
}
