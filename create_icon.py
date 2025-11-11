#!/usr/bin/env python3
"""
Create a scanner icon for the system tray.
Generates a simple scanner icon in ICO format with multiple sizes.
"""

from PIL import Image, ImageDraw

def create_scanner_icon(size):
    """Create a scanner icon of the specified size."""
    # Create a new image with transparent background
    img = Image.new('RGBA', (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(img)

    # Calculate dimensions based on size
    padding = max(2, size // 8)

    # Scanner body (rectangle)
    body_top = padding
    body_bottom = size - padding
    body_left = padding
    body_right = size - padding

    # Draw scanner body outline
    draw.rectangle(
        [(body_left, body_top), (body_right, body_bottom)],
        outline=(70, 70, 70, 255),
        width=max(1, size // 16)
    )

    # Fill scanner body with light gray
    fill_padding = max(1, size // 16)
    draw.rectangle(
        [(body_left + fill_padding, body_top + fill_padding),
         (body_right - fill_padding, body_bottom - fill_padding)],
        fill=(200, 200, 200, 255)
    )

    # Scanner lid (darker top section)
    lid_height = max(3, size // 4)
    draw.rectangle(
        [(body_left + fill_padding, body_top + fill_padding),
         (body_right - fill_padding, body_top + lid_height)],
        fill=(120, 120, 120, 255)
    )

    # Document being scanned (white rectangle)
    doc_padding = max(2, size // 5)
    doc_top = body_top + lid_height + max(2, size // 16)
    doc_bottom = body_bottom - max(2, size // 8)
    doc_left = body_left + doc_padding
    doc_right = body_right - doc_padding

    # Draw white document
    draw.rectangle(
        [(doc_left, doc_top), (doc_right, doc_bottom)],
        fill=(255, 255, 255, 255),
        outline=(100, 100, 100, 255),
        width=1
    )

    # Add scan lines on the document
    num_lines = max(2, size // 8)
    line_spacing = (doc_bottom - doc_top) / (num_lines + 1)
    for i in range(1, num_lines + 1):
        y = int(doc_top + i * line_spacing)
        line_left = doc_left + max(1, size // 16)
        line_right = doc_right - max(1, size // 16)
        draw.line(
            [(line_left, y), (line_right, y)],
            fill=(180, 180, 180, 255),
            width=1
        )

    # Add a scanning light effect (cyan/blue line)
    scan_line_y = int(doc_top + (doc_bottom - doc_top) * 0.4)
    draw.line(
        [(doc_left, scan_line_y), (doc_right, scan_line_y)],
        fill=(0, 200, 255, 200),
        width=max(1, size // 12)
    )

    return img

def main():
    """Generate the scanner icon in ICO format."""
    # Create icons at multiple sizes (required for ICO format)
    sizes = [16, 32, 48, 64, 128, 256]
    images = [create_scanner_icon(size) for size in sizes]

    # Save as ICO file
    output_file = "scanner_icon.ico"
    images[0].save(
        output_file,
        format='ICO',
        sizes=[(img.width, img.height) for img in images],
        append_images=images[1:]
    )

    print(f"Created scanner icon: {output_file}")
    print(f"  Sizes included: {', '.join(f'{s}x{s}' for s in sizes)}")

    # Also save a PNG version for preview
    preview_size = 256
    preview = create_scanner_icon(preview_size)
    preview.save("scanner_icon_preview.png")
    print(f"Created preview: scanner_icon_preview.png ({preview_size}x{preview_size})")

if __name__ == '__main__':
    main()
