/*
 * Copyright (c) 2015, Robert Bieber
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 *
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 *
 * 2. Redistributions in binary form must reproduce the above
 *    copyright notice, this list of conditions and the following
 *    disclaimer in the documentation and/or other materials provided
 *    with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS
 * FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
 * COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
 * INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
 * HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
 * STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED
 * OF THE POSSIBILITY OF SUCH DAMAGE.
 *
 */

// barcode provides a wrapper around zbar for reading barcodes from
// images.  For the time being, it only supports reading from images
// and not any of zbar's more esoteric features like GUI display and
// reading directly from barcode scanners or video inputs.  I probably
// won't implement any further features as I don't have any need for
// them, but feel free to submit a pull request if you need them and
// want to wrap them yourself.
//
// Because this package uses cgo, it links its dependencies
// dynamically by default.  To compile you'll need the zbar C library
// and headers installed, and you'll need the library installed to run
// binaries compiled with this package.
//
// To read codes from an image, simply instantiate an Image (you can
// create one from any image.Image) and then pass it to an
// ImageScanner.  For example:
//
//	import (
//		"fmt"
//		"gopkg.in/bieber/barcode.v0"
//		"image/jpeg"
//		"os"
//	)
//
//	func main() {
//		fin, _ := os.Open("card.jpg")
//		defer fin.Close()
//		src, _ := jpeg.Decode(fin)
//
//		img := barcode.NewImage(src)
//		scanner := barcode.NewScanner().
//			SetEnabledAll(true)
//
//		symbols, _ := scanner.ScanImage(img)
//		for _, s := range symbols {
//			fmt.Println(s.Type.Name(), s.Data, s.Quality, s.Boundary)
//		}
//	}
//
package barcode
