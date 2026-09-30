#!/usr/bin/env python3
"""Regenerate the three printable ArUco markers (DICT_4X4_50, ids 0–2, 800 px)."""

import zlib
from pathlib import Path

import cv2

ROOT = Path(__file__).resolve().parent
EDGE_PX = 800
NOTES = [
    "Marker {n} of 3. Print all three and stick them on different sides of what you are wrapping.",
    "Keep them flat and do not move them between clips.",
    "Each clip should show at least one marker. To join two sides that do not share a marker,",
    "at least one clip must show two markers at the same time.",
]


def _escape(text: str) -> str:
    return text.replace("\\", "\\\\").replace("(", "\\(").replace(")", "\\)")


def _pdf(image: bytes, width: int, height: int, lines: list[str]) -> bytes:
    content = "q 283.46 0 0 283.46 155.77 438.54 cm /Im0 Do Q\nBT /F1 11 Tf 50 390 Td 14 TL\n"
    for i, line in enumerate(lines):
        if i:
            content += "T* "
        content += f"({_escape(line)}) Tj\n"
    content += "ET\n"
    content_b = content.encode("latin1")
    objects = [
        b"<< /Type /Catalog /Pages 2 0 R >>",
        b"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
        b"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] "
        b"/Resources << /XObject << /Im0 5 0 R >> /Font << /F1 6 0 R >> >> "
        b"/Contents 4 0 R >>",
        b"<< /Length %d >>\nstream\n%s\nendstream" % (len(content_b), content_b),
        b"<< /Type /XObject /Subtype /Image /Width %d /Height %d "
        b"/ColorSpace /DeviceGray /BitsPerComponent 8 /Filter /FlateDecode /Length %d >>\nstream\n"
        % (width, height, len(image))
        + image
        + b"\nendstream",
        b"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
    ]
    out = bytearray(b"%PDF-1.4\n")
    offsets = []
    for i, body in enumerate(objects, start=1):
        offsets.append(len(out))
        out += f"{i} 0 obj\n".encode()
        out += body
        out += b"\nendobj\n"
    xref = len(out)
    out += f"xref\n0 {len(objects) + 1}\n".encode()
    out += b"0000000000 65535 f \n"
    for off in offsets:
        out += f"{off:010d} 00000 n \n".encode()
    out += (
        f"trailer\n<< /Size {len(objects) + 1} /Root 1 0 R >>\nstartxref\n{xref}\n%%EOF\n"
    ).encode()
    return bytes(out)


def main() -> None:
    dictionary = cv2.aruco.getPredefinedDictionary(cv2.aruco.DICT_4X4_50)
    for marker_id in (0, 1, 2):
        image = cv2.aruco.generateImageMarker(dictionary, marker_id, EDGE_PX)
        stem = f"fiducial_marker_aruco4x4_50_id{marker_id}_100mm"
        cv2.imwrite(str(ROOT / f"{stem}.png"), image)
        raw = image.tobytes()
        lines = [line.format(n=marker_id + 1) for line in NOTES]
        (ROOT / f"{stem}.pdf").write_bytes(_pdf(zlib.compress(raw), EDGE_PX, EDGE_PX, lines))


if __name__ == "__main__":
    main()
