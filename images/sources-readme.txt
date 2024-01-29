
Due to the way this image is built, the source container for the image itself
does not include the full set of sources for the cli tool binaries included in
the image, i.e. `cosign`, `gitsign`, `rekor-cli` and `ec`.

The complete sources for each cli tool are available in the source containers
for the images that the binaries are extracted from.
