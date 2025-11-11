#!/usr/bin/env python3
"""
Generate a test ICC profile that increases red by 22 (out of 255).
This creates a simple RGB matrix-based ICC profile with custom tone curves.
"""

import struct
import datetime

def create_curve_data(offset=0):
    """
    Create a curve (gamma) table with an offset applied.
    For red channel, offset=22 to increase red values.
    For green and blue, offset=0 to keep them unchanged.

    Returns a list of 256 values representing the tone curve.
    """
    curve = []
    for i in range(256):
        # Add offset and clamp to valid range [0, 255]
        value = min(255, max(0, i + offset))
        # Convert to 16-bit (ICC uses 16-bit values)
        value_16bit = int((value / 255.0) * 65535)
        curve.append(value_16bit)
    return curve

def write_tag(tag_signature, data):
    """Helper to create an ICC tag with signature and data."""
    return tag_signature + struct.pack('>I', 0) + data

def pad_to_4bytes(data):
    """Pad data to 4-byte boundary."""
    remainder = len(data) % 4
    if remainder:
        data += b'\x00' * (4 - remainder)
    return data

def create_icc_profile():
    """Create a complete ICC profile that increases red by 22."""

    # Profile header (128 bytes)
    profile_size_placeholder = b'\x00\x00\x00\x00'  # Will be updated at the end
    preferred_cmm = b'\x00\x00\x00\x00'  # No preferred CMM
    profile_version = b'\x02\x10\x00\x00'  # Version 2.1
    profile_class = b'mntr'  # Monitor/Display profile
    color_space = b'RGB '  # RGB color space
    pcs = b'XYZ '  # Profile Connection Space (XYZ)
    creation_date = b'\x07\xe9\x0b\x0a\x00\x00\x00\x00\x00\x00\x00\x00'  # Date/Time
    acsp = b'acsp'  # Profile file signature
    platform = b'MSFT'  # Microsoft platform
    flags = b'\x00\x00\x00\x00'
    device_manufacturer = b'\x00\x00\x00\x00'
    device_model = b'\x00\x00\x00\x00'
    device_attributes = b'\x00\x00\x00\x00\x00\x00\x00\x00'
    rendering_intent = b'\x00\x00\x00\x00'  # Perceptual
    pcs_illuminant = struct.pack('>III', 0x0000F6D6, 0x00010000, 0x0000D32D)  # D50
    creator = b'CLDE'  # Claude Code
    reserved = b'\x00' * 44

    header = (profile_size_placeholder + preferred_cmm + profile_version +
              profile_class + color_space + pcs + creation_date + acsp +
              platform + flags + device_manufacturer + device_model +
              device_attributes + rendering_intent + pcs_illuminant +
              creator + reserved)

    # Tag table
    tag_count = 9  # Number of tags we'll include

    # Create curve data for each channel
    red_curve = create_curve_data(offset=22)    # Increase red by 22
    green_curve = create_curve_data(offset=0)   # No change
    blue_curve = create_curve_data(offset=0)    # No change

    # Curve tag type signature and data
    def create_curve_tag(curve_values):
        """Create a curveType tag."""
        curve_type = b'curv'
        reserved = b'\x00\x00\x00\x00'
        count = struct.pack('>I', len(curve_values))
        values = b''.join(struct.pack('>H', v) for v in curve_values)
        return curve_type + reserved + count + values

    red_curve_tag = create_curve_tag(red_curve)
    green_curve_tag = create_curve_tag(green_curve)
    blue_curve_tag = create_curve_tag(blue_curve)

    # Text description tag
    def create_text_desc_tag(text):
        """Create a textDescriptionType tag."""
        tag_type = b'desc'
        reserved = b'\x00\x00\x00\x00'
        ascii_len = struct.pack('>I', len(text) + 1)
        ascii_text = text.encode('ascii') + b'\x00'
        unicode_code = struct.pack('>I', 0)  # No Unicode
        unicode_len = struct.pack('>I', 0)
        script_code = struct.pack('>H', 0)  # No ScriptCode
        script_len = b'\x00'
        return (tag_type + reserved + ascii_len + ascii_text +
                unicode_code + unicode_len + script_code + script_len)

    desc_tag = create_text_desc_tag("Red+22 Test Profile")
    copyright_tag = create_text_desc_tag("Public Domain")

    # XYZ tags for colorimetry (standard D50 values)
    def create_xyz_tag(x, y, z):
        """Create an XYZType tag."""
        tag_type = b'XYZ '
        reserved = b'\x00\x00\x00\x00'
        return tag_type + reserved + struct.pack('>III', x, y, z)

    # D50 white point
    wtpt_tag = create_xyz_tag(0x0000F6D6, 0x00010000, 0x0000D32D)

    # RGB primaries (standard sRGB-like values)
    rXYZ_tag = create_xyz_tag(0x0000F351, 0x00010000, 0x0000D32D)
    gXYZ_tag = create_xyz_tag(0x00006FA2, 0x00010000, 0x0000D32D)
    bXYZ_tag = create_xyz_tag(0x00006FA2, 0x00010000, 0x0000D32D)

    # Assemble tags with offsets
    # Tag table starts at offset 128
    tags_data = []
    current_offset = 128 + 4 + (tag_count * 12)  # After header and tag table

    tag_signatures = [
        (b'desc', desc_tag),
        (b'cprt', copyright_tag),
        (b'wtpt', wtpt_tag),
        (b'rXYZ', rXYZ_tag),
        (b'gXYZ', gXYZ_tag),
        (b'bXYZ', bXYZ_tag),
        (b'rTRC', red_curve_tag),
        (b'gTRC', green_curve_tag),
        (b'bTRC', blue_curve_tag),
    ]

    tag_table_entries = []
    tag_data_section = b''

    for sig, tag_data in tag_signatures:
        # Pad tag data to 4-byte boundary
        tag_data = pad_to_4bytes(tag_data)

        # Add to tag table
        tag_table_entries.append(
            sig +
            struct.pack('>I', current_offset) +
            struct.pack('>I', len(tag_data))
        )

        # Add to data section
        tag_data_section += tag_data
        current_offset += len(tag_data)

    # Create tag table
    tag_table = struct.pack('>I', tag_count) + b''.join(tag_table_entries)

    # Combine all parts
    profile = header + tag_table + tag_data_section

    # Update profile size in header
    profile_size = len(profile)
    profile = struct.pack('>I', profile_size) + profile[4:]

    return profile

def main():
    """Generate the ICC profile and save to file."""
    profile_data = create_icc_profile()

    output_file = "red_plus_22.icc"
    with open(output_file, 'wb') as f:
        f.write(profile_data)

    print(f"Generated ICC profile: {output_file}")
    print(f"  Profile size: {len(profile_data)} bytes")
    print(f"  Effect: Increases red channel by 22 (out of 255)")
    print(f"\nUsage:")
    print(f"  scanner.exe -process-file <image.tiff> -profile {output_file}")

if __name__ == '__main__':
    main()
