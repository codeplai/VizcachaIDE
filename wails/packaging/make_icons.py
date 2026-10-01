"""Build every VizcachaIDE icon from the brand logo (vizcachaidelogo.png at the repo root).

Outputs (all regenerated, commit the results):
  wails/build/appicon.png                    1024 px, mascot only (Wails derives the macOS icon from
  it)
  wails/build/windows/icon.ico               16-256 px; 256 px shows the full logo, smaller sizes
  the mascot
  wails/frontend/src/assets/brand/mark.png   128 px mascot for the title bar
  wails/frontend/src/assets/brand/logo.png   full logo for About and the first-run screen

Usage (needs Pillow, dev only):  python wails/packaging/make_icons.py
"""

from pathlib import Path

from PIL import Image

REPO_ROOT = Path(__file__).resolve().parents[2]
SOURCE = REPO_ROOT / "vizcachaidelogo.png"
WAILS = REPO_ROOT / "wails"

# The mascot badge, above the "VizcachaIDE by codeplai." wordmark (measured on the 1536x1024
# source).
MASCOT_BOX = (334, 36, 1203, 700)
ICO_MASCOT_SIZES = (16, 24, 32, 48, 64, 128)
FULL_LOGO_WIDTH = 480


def square(image: Image.Image, padding_ratio: float = 0.04) -> Image.Image:
    """Center ``image`` on a transparent square canvas with a small margin."""
    side = round(max(image.size) * (1 + padding_ratio * 2))
    canvas = Image.new("RGBA", (side, side), (0, 0, 0, 0))
    canvas.paste(image, ((side - image.width) // 2, (side - image.height) // 2), image)
    return canvas


def trimmed(image: Image.Image) -> Image.Image:
    """Crop away fully transparent borders."""
    box = image.getchannel("A").point(lambda alpha: 255 if alpha > 8 else 0).getbbox()
    return image.crop(box) if box else image


def resized(image: Image.Image, side: int) -> Image.Image:
    return image.resize((side, side), Image.Resampling.LANCZOS)


def build(source: Path = SOURCE) -> None:
    logo = Image.open(source).convert("RGBA")
    mascot = square(trimmed(logo.crop(MASCOT_BOX)))
    full = square(trimmed(logo))

    resized(mascot, 1024).save(WAILS / "build" / "appicon.png", optimize=True)

    icon_images = [resized(full, 256)] + [resized(mascot, size) for size in ICO_MASCOT_SIZES]
    icon_images[0].save(
        WAILS / "build" / "windows" / "icon.ico",
        format="ICO",
        sizes=[(image.width, image.height) for image in icon_images],
        append_images=icon_images[1:],
    )

    brand = WAILS / "frontend" / "src" / "assets" / "brand"
    brand.mkdir(parents=True, exist_ok=True)
    wordmark = trimmed(logo)
    height = round(wordmark.height * FULL_LOGO_WIDTH / wordmark.width)
    small = wordmark.resize((FULL_LOGO_WIDTH, height), Image.Resampling.LANCZOS)
    # 256-colour palette keeps the transparency and makes the in-app asset ~5x smaller.
    small.quantize(colors=256, method=Image.Quantize.FASTOCTREE).save(
        brand / "logo.png", optimize=True
    )
    resized(mascot, 128).quantize(colors=256, method=Image.Quantize.FASTOCTREE).save(
        brand / "mark.png", optimize=True
    )


if __name__ == "__main__":
    build()
    print("icons written")
