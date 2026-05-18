# mtex-render

`mtex-render` is a simple command that renders a LaTeX equation to a PDF or PNG document:

```
$> mtex-render -h
Usage of mtex-render:
  -dpi float
    	dots-per-inch to use (default 72)
  -font-size float
    	font size to use (default 12)
  -o string
    	path to output file (default "out.png")

$> mtex-render -dpi 250 -font-size 42 -o foo.png "$\frac{2\pi}{\sqrt{x+\partial x}}$"
```

![img-cli](https://codeberg.org/go-latex/latex/raw/main/cmd/mtex-render/testdata/cli.png)
