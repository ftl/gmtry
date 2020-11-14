package gmtry

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/ftl/gmtry/pb"
)

// ID of a window.
type ID string

func NewGeometry(filename string) *Geometry {
	return &Geometry{
		filename:     filename,
		connectables: NewConnectables(),
		windows:      NewWindows(),
	}
}

type Geometry struct {
	filename     string
	connectables Connectables
	windows      Windows
}

func (g *Geometry) Add(id ID, connectable Connectable) {
	var window *Window
	window, ok := g.windows[id]
	if ok {
		window.Apply(connectable)
	} else {
		window = &Window{ID: id}
		window.SetPosition(connectable.GetPosition())
		window.SetSize(connectable.GetSize())
		window.SetMaximized(connectable.IsMaximized())
		g.windows[id] = window
	}
	g.connectables[id] = connectable
}

func (g *Geometry) Get(id ID) *Window {
	return g.windows.Get(id)
}

func (g *Geometry) Store() error {
	f, err := os.OpenFile(g.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("Cannot open window geometry file %s: %w", g.filename, err)
	}
	defer f.Close()

	err = g.windows.Store(f)
	if err != nil {
		return fmt.Errorf("Cannot store window geometry: %w", err)
	}

	log.Printf("Stored window geometry in %s", f.Name())
	return nil
}

func (g *Geometry) Restore() error {
	log.Printf("Loading window geometry from %s", g.filename)

	f, err := os.Open(g.filename)
	if err != nil {
		return fmt.Errorf("Cannot open %s: %w", g.filename, err)
	}
	defer f.Close()

	loaded, err := LoadWindows(f)
	if err != nil {
		return fmt.Errorf("Cannot load window geometry: %w", err)
	}

	for id, window := range loaded {
		g.windows[id] = window
		connectable, ok := g.connectables[id]
		if !ok {
			continue
		}
		window.Apply(connectable)
	}
	return nil
}

// Window contains all data about a window.
type Window struct {
	ID        ID
	X         int
	Y         int
	Width     int
	Height    int
	Maximized bool
}

func (w *Window) String() string {
	return fmt.Sprintf("Window %s: (%d, %d) (%d x %d) %t", w.ID, w.X, w.Y, w.Width, w.Height, w.Maximized)
}

// Apply the window geometry to the given target.
func (w *Window) Apply(a Applyable) {
	a.Move(w.X, w.Y)
	a.Resize(w.Width, w.Height)
	if w.Maximized {
		a.Maximize()
	}
}

// SetPosition of this window.
func (w *Window) SetPosition(x, y int) {
	if w.Maximized {
		return
	}
	w.X = x
	w.Y = y
}

// SetSize of this window.
func (w *Window) SetSize(width, height int) {
	if w.Maximized {
		return
	}
	w.Width = width
	w.Height = height
}

// SetMaximized flag of this window.
func (w *Window) SetMaximized(maximized bool) {
	w.Maximized = maximized
}

// Applyable represents anything that window geometry can be applied to.
type Applyable interface {
	Move(x, y int)
	Resize(width, height int)
	Maximize()
}

// Observable represents anything that window geometry can be retireved from.
type Observable interface {
	GetPosition() (x, y int)
	GetSize() (width, height int)
	IsMaximized() bool
}

// Connectable describes anything that is both an Applyable and an Observable.
type Connectable interface {
	Applyable
	Observable
}

// Connectables contains connectables mapped by their ID.
type Connectables map[ID]Connectable

// NewConnectables instance.
func NewConnectables() Connectables {
	return make(Connectables)
}

// Windows contains windows mapped by their ID.
type Windows map[ID]*Window

// NewWindows instance.
func NewWindows() Windows {
	return make(Windows)
}

// LoadWindows from the given reader.
func LoadWindows(r io.Reader) (Windows, error) {
	buffer, err := ioutil.ReadAll(r)
	if err != nil {
		return NewWindows(), err
	}
	pbWindows := new(pb.Windows)
	err = proto.Unmarshal(buffer, pbWindows)
	if err != nil {
		return NewWindows(), err
	}
	result := NewWindows()
	for _, pbWindow := range pbWindows.Windows {
		window := Window{
			ID:        ID(pbWindow.Name),
			X:         int(pbWindow.Position.X),
			Y:         int(pbWindow.Position.Y),
			Width:     int(pbWindow.Size.Width),
			Height:    int(pbWindow.Size.Height),
			Maximized: pbWindow.Maximized,
		}
		result[window.ID] = &window
	}
	return result, nil
}

// Store windows to the given writer.
func (w Windows) Store(writer io.Writer) error {
	pbWindows := new(pb.Windows)
	for _, window := range w {
		pbWindow := pb.Window{
			Name:      string(window.ID),
			Position:  &pb.Position{X: int32(window.X), Y: int32(window.Y)},
			Size:      &pb.Size{Width: int32(window.Width), Height: int32(window.Height)},
			Maximized: window.Maximized,
		}
		pbWindows.Windows = append(pbWindows.Windows, &pbWindow)
	}
	bytes, err := proto.Marshal(pbWindows)
	if err != nil {
		return errors.Wrap(err, "cannot marshal the windows")
	}
	n, err := writer.Write(bytes)
	if err != nil {
		return errors.Wrap(err, "cannot write windows")
	}
	if n != len(bytes) {
		return errors.Errorf("could only write %d of %d bytes", n, len(bytes))
	}
	return nil
}

// Get the window with the given ID.
func (w Windows) Get(id ID) *Window {
	g, ok := w[id]
	if !ok {
		g = &Window{
			ID: id,
		}
		w[id] = g
	}
	return g
}

func (w Windows) String() string {
	result := ""
	for _, window := range w {
		result += fmt.Sprintf("%v\n", window)
	}
	return result
}
