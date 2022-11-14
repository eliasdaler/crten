# crten v0.1 - Cathode-Ray Tube ENgine
Display/render pixel art with a CRT effect.

## Web demo / binaries download

https://eliasdaler.itch.io/crten

Written with [Ebitengine](https://github.com/hajimehoshi/ebiten).

![image](https://user-images.githubusercontent.com/1285136/201767097-80f0a4d8-b8e4-4763-9db1-fbc88f57e5ba.png)

![image](https://user-images.githubusercontent.com/1285136/201767852-e3247b2b-81d6-4072-9244-bcc02ceeaa38.png)
"Mountais of Madness" by [Polyducks](http://polyducks.co.uk/)


## Usage

`crten IMAGE_FILE` - display INPUT_FILE with CRT effect.

`crten -i INPUT_FILE [-c CONFIG_FILE] OUTPUT_FILE` - renders INPUT_FILE image with CRT effect to OUTPUT_FILE and closes the window.


## License

* assets/*.png - used in web demo, with respective authors permission. See `assets/metadata.json` for art details.
* [m5x7 font](https://managore.itch.io/m5x7) by [Daniel Linsenn](https://twitter.com/managore) - "free to use but attribution appreciated"
* crt-lottes - a port of public domain shader by Timothy Lottes. The license for it is still public domain. See the [original source](https://github.com/libretro/glsl-shaders/blob/master/crt/shaders/crt-lottes.glsl)
* the program itself - GPLv3
