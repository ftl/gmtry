package gmtry

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundtrip(t *testing.T) {
	initials := NewWindows()
	main := ID("main")
	initialMain := initials.Get(main)
	initialMain.SetPosition(1, 2)
	initialMain.SetSize(3, 4)
	initialMain.SetMaximized(true)
	dialog := ID("dialog")
	initialDialog := initials.Get(dialog)
	initialDialog.SetPosition(10, 20)
	initialDialog.SetSize(300, 400)
	initialDialog.SetMaximized(false)

	buffer := bytes.NewBuffer([]byte{})
	initials.Store(buffer)

	loaded, err := LoadWindows(buffer)
	require.NoError(t, err)

	assert.Equal(t, initialMain, loaded.Get(main))
	assert.Equal(t, initialDialog, loaded.Get(dialog))
}

func TestRoundtripWithFile(t *testing.T) {
	initials := NewWindows()
	main := ID("main")
	initialMain := initials.Get(main)
	initialMain.SetPosition(1, 2)
	initialMain.SetSize(3, 4)
	initialMain.SetMaximized(true)
	dialog := ID("dialog")
	initialDialog := initials.Get(dialog)
	initialDialog.SetPosition(10, 20)
	initialDialog.SetSize(300, 400)
	initialDialog.SetMaximized(false)

	writeFile, err := ioutil.TempFile("", "TestRoundtripWithFile")
	require.NoError(t, err)
	defer writeFile.Close()

	initials.Store(writeFile)

	readFile, err := os.Open(writeFile.Name())
	require.NoError(t, err)
	defer readFile.Close()

	loaded, err := LoadWindows(readFile)
	require.NoError(t, err)

	assert.Equal(t, initialMain, loaded.Get(main))
	assert.Equal(t, initialDialog, loaded.Get(dialog))
}

func TestAdd_ShouldStoreConnectableGeometry(t *testing.T) {
	c := &testConnectable{
		Window: Window{
			ID:     "c",
			X:      10,
			Y:      20,
			Width:  100,
			Height: 200,
		},
	}
	g := NewGeometry("")
	g.Add(c.ID, c)
	w := g.Get("c")

	assert.Equal(t, c.Window, *w)
}

func TestAddAgain_ShouldRestoreGeometryOnConnectable(t *testing.T) {
	c1 := &testConnectable{
		Window: Window{
			X:      10,
			Y:      20,
			Width:  100,
			Height: 200,
		},
	}
	c2 := new(testConnectable)
	g := NewGeometry("")
	g.Add("c", c1)
	g.Add("c", c2)
	assert.Equal(t, c1, c2)
}

type testConnectable struct {
	Window
}

func (c *testConnectable) Move(x, y int) {
	c.X, c.Y = x, y
}

func (c *testConnectable) Resize(width, height int) {
	c.Width, c.Height = width, height
}

func (c *testConnectable) Maximize() {
	c.Maximized = true
}

func (c *testConnectable) GetPosition() (x, y int) {
	return c.X, c.Y
}

func (c *testConnectable) GetSize() (width, height int) {
	return c.Width, c.Height
}

func (c *testConnectable) IsMaximized() bool {
	return c.Maximized
}
