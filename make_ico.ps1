Add-Type -AssemblyName System.Drawing

$srcPath = Join-Path $PSScriptRoot 'charsaywhat.png'
$dstPath = Join-Path $PSScriptRoot 'charsaywhat.ico'

$src = [System.Drawing.Image]::FromFile($srcPath)
$sizes = @(16, 24, 32, 48, 64, 128, 256)

$pngBlobs = @()
foreach ($size in $sizes) {
    $bmp = New-Object System.Drawing.Bitmap $size, $size
    $g = [System.Drawing.Graphics]::FromImage($bmp)
    $g.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
    $g.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::HighQuality
    $g.PixelOffsetMode = [System.Drawing.Drawing2D.PixelOffsetMode]::HighQuality
    $g.CompositingQuality = [System.Drawing.Drawing2D.CompositingQuality]::HighQuality
    $g.DrawImage($src, 0, 0, $size, $size)
    $g.Dispose()

    $ms = New-Object System.IO.MemoryStream
    $bmp.Save($ms, [System.Drawing.Imaging.ImageFormat]::Png)
    $bmp.Dispose()
    $pngBlobs += ,($ms.ToArray())
    $ms.Dispose()
}
$src.Dispose()

$out = New-Object System.IO.MemoryStream
$bw = New-Object System.IO.BinaryWriter($out)

# ICONDIR
$bw.Write([uint16]0)              # reserved
$bw.Write([uint16]1)              # type = icon
$bw.Write([uint16]$sizes.Count)   # count

$headerSize = 6 + (16 * $sizes.Count)
$offset = $headerSize

# ICONDIRENTRY per image
for ($i = 0; $i -lt $sizes.Count; $i++) {
    $s = $sizes[$i]
    $blob = $pngBlobs[$i]
    $dim = if ($s -ge 256) { 0 } else { [byte]$s }
    $bw.Write([byte]$dim)          # width
    $bw.Write([byte]$dim)          # height
    $bw.Write([byte]0)             # palette
    $bw.Write([byte]0)             # reserved
    $bw.Write([uint16]1)           # planes
    $bw.Write([uint16]32)          # bpp
    $bw.Write([uint32]$blob.Length)
    $bw.Write([uint32]$offset)
    $offset += $blob.Length
}

# Image data
foreach ($blob in $pngBlobs) {
    $bw.Write($blob)
}

[System.IO.File]::WriteAllBytes($dstPath, $out.ToArray())
$bw.Dispose()
$out.Dispose()

Write-Output "Wrote $dstPath ($((Get-Item $dstPath).Length) bytes)"
