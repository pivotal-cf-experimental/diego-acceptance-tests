package assets

type Assets struct {
	Dora       string
	HelloWorld string
	Standalone string
	Fuse       string
}

func NewAssets() Assets {
	return Assets{
		Dora:       "../assets/dora",
		HelloWorld: "../assets/hello-world",
		Standalone: "../assets/standalone",
		Fuse:       "../assets/fuse-mount",
	}
}
