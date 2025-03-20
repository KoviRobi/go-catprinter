package catprinter

import "fmt"

var (
	cmdGetDevState      = []byte{81, 120, 163, 0, 1, 0, 0, 0, 255}
	cmdSetQuality200Dpi = []byte{81, 120, 164, 0, 1, 0, 50, 158, 255}
	cmdGetDevInfo       = []byte{81, 120, 168, 0, 1, 0, 0, 0, 255}
	cmdLatticeStart     = []byte{81, 120, 166, 0, 11, 0, 170, 85, 23, 56, 68, 95, 95, 95, 68, 56, 44, 161, 255}
	cmdLatticeEnd       = []byte{81, 120, 166, 0, 11, 0, 170, 85, 23, 0, 0, 0, 0, 0, 0, 0, 23, 17, 255}
	cmdSetPaper         = []byte{81, 120, 161, 0, 2, 0, 48, 0, 249, 255}
	cmdPrintImg         = []byte{81, 120, 190, 0, 1, 0, 0, 0, 255}
	cmdPrintText        = []byte{81, 120, 190, 0, 1, 0, 1, 7, 255}

	cmdStartPrinting = []byte{0x51, 0x78, 0xa3, 0x00, 0x01, 0x00, 0x00, 0x00, 0xff}
	cmdApplyEnergy   = []byte{81, 120, 190, 0, 1, 0, 1, 7, 255}
	cmdUpdateDevice  = []byte{81, 120, 169, 0, 1, 0, 0, 0, 255}
	cmdSlow          = []byte{81, 120, 189, 0, 1, 0, 36, 252, 255}
	cmdPower         = []byte{81, 120, 175, 0, 2, 0, 255, 223, 196, 255}

	cmdFinalSpeed = []byte{81, 120, 189, 0, 1, 0, 8, 56, 255}

	checksumTable = []byte{
		0, 7, 14, 9, 28, 27, 18, 21, 56, 63, 54, 49, 36, 35, 42, 45, 112, 119, 126, 121, 108, 107, 98,
		101, 72, 79, 70, 65, 84, 83, 90, 93, 224, 231, 238, 233, 252, 251, 242, 245, 216, 223, 214, 209, 196, 195, 202,
		205, 144, 151, 158, 153, 140, 139, 130, 133, 168, 175, 166, 161, 180, 179, 186, 189, 199, 192, 201, 206, 219, 220,
		213, 210, 255, 248, 241, 246, 227, 228, 237, 234, 183, 176, 185, 190, 171, 172, 165, 162, 143, 136, 129, 134, 147,
		148, 157, 154, 39, 32, 41, 46, 59, 60, 53, 50, 31, 24, 17, 22, 3, 4, 13, 10, 87, 80, 89, 94, 75, 76, 69, 66, 111,
		104, 97, 102, 115, 116, 125, 122, 137, 142, 135, 128, 149, 146, 155, 156, 177, 182, 191, 184, 173, 170, 163, 164,
		249, 254, 247, 240, 229, 226, 235, 236, 193, 198, 207, 200, 221, 218, 211, 212, 105, 110, 103, 96, 117, 114, 123,
		124, 81, 86, 95, 88, 77, 74, 67, 68, 25, 30, 23, 16, 5, 2, 11, 12, 33, 38, 47, 40, 61, 58, 51, 52, 78, 73, 64, 71,
		82, 85, 92, 91, 118, 113, 120, 127, 106, 109, 100, 99, 62, 57, 48, 55, 34, 37, 44, 43, 6, 1, 8, 15, 26, 29, 20, 19,
		174, 169, 160, 167, 178, 181, 188, 187, 150, 145, 152, 159, 138, 141, 132, 131, 222, 217, 208, 215, 194, 197, 204,
		203, 230, 225, 232, 239, 250, 253, 244, 243,
	}
)

// checksum calculates the checksum for a given byte array.
func checksum(bArr []byte, i, i2 int) byte {
	var b2 byte
	for i3 := i; i3 < i+i2; i3++ {
		b2 = checksumTable[(b2^bArr[i3])&0xff]
	}
	return b2
}

// commandRetractPaper creates a command to retract paper by a specified amount.
func commandRetractPaper(howMuch int) []byte {

	bArr := []byte{
		81,
		120,
		160,
		0,
		1,
		0,
		byte(howMuch & 0xff),
		0,
		255,
	}

	bArr[7] = checksum(bArr, 6, 1)
	return bArr
}

// commandFeedPaper creates a command to feed paper by a specified amount.
func commandFeedPaper(howMuch int) []byte {

	bArr := []byte{
		81,
		120,
		161,
		0,
		1,
		0,
		byte(howMuch & 0xff),
		0,
		255,
	}

	bArr[7] = checksum(bArr, 6, 1)
	return bArr
}

// cmdSetEnergy sets the energy level. Max to `0xffff` By default, it seems around `0x3000` (1 / 5)
func commandSetEnergy(val int) []byte {

	bArr := []byte{
		81,
		120,
		175,
		0,
		2,
		0,
		byte((val >> 8) & 0xff),
		byte(val & 0xff),
		0,
		255,
	}

	bArr[7] = checksum(bArr, 6, 2)
	fmt.Println(bArr)
	return bArr

}

// encodeRunLengthRepetition encodes repetitions in a run-length format.
func encodeRunLengthRepetition(n int, val byte) []byte {
	var res []byte
	for n > 0x7f {
		res = append(res, 0x7f|byte(val<<7))
		n -= 0x7f
	}
	if n > 0 {
		res = append(res, val<<7|byte(n))
	}
	return res
}

// runLengthEncode performs run-length encoding on an image row.
func runLengthEncode(imgRow []byte) []byte {
	var res []byte
	count := 0
	var lastVal byte = 255
	for _, val := range imgRow {
		if val == lastVal {
			count++
		} else {
			res = append(res, encodeRunLengthRepetition(count, lastVal)...)
			count = 1
		}
		lastVal = val
	}
	if count > 0 {
		res = append(res, encodeRunLengthRepetition(count, lastVal)...)
	}
	return res
}

// byteEncode encodes an image row into a byte array.
func byteEncode(imgRow []byte) []byte {
	var res []byte
	for chunkStart := 0; chunkStart < len(imgRow); chunkStart += 8 {
		var byteVal byte = 0
		for bitIndex := 0; bitIndex < 8; bitIndex++ {
			if chunkStart+bitIndex < len(imgRow) && imgRow[chunkStart+bitIndex] != 0 {
				byteVal |= 1 << uint8(bitIndex)
			}
		}
		res = append(res, byteVal)
	}
	return res
}

// commandPrintRow builds a print row command based on the image data.
func commandPrintRow(imgRow []byte) []byte {

	// Try to use run-length compression on the image data.
	encodedImg := runLengthEncode(imgRow)

	// If the resulting compression takes more than PRINT_WIDTH // 8, it means
	// it's not worth it. So we fall back to a simpler, fixed-length encoding.
	if len(encodedImg) > printWidth/8 {
		encodedImg = byteEncode(imgRow)
		bArr := append([]byte{
			81,
			120,
			162,
			0,
			byte(len(encodedImg) & 0xff),
			0,
		}, encodedImg...)
		bArr = append(bArr, 0, 0xff)
		bArr[len(bArr)-2] = checksum(bArr, 6, len(encodedImg))
		return bArr
	}

	// Build the run-length encoded image command.
	bArr := append([]byte{81, 120, 191, 0, byte(len(encodedImg)), 0}, encodedImg...)
	bArr = append(bArr, 0, 0xff)
	bArr[len(bArr)-2] = checksum(bArr, 6, len(encodedImg))
	return bArr
}

// commandsPrintImg builds the commands to print an image.
func commandsPrintImg(imgS []byte, feed int) []byte {

	img := chunkify(imgS, printWidth)
	var data []byte

	data = append(data, cmdGetDevState...)
	data = append(data, cmdStartPrinting...)
	data = append(data, cmdSetQuality200Dpi...)
	data = append(data, cmdSlow...)
	data = append(data, cmdPower...)
	data = append(data, cmdApplyEnergy...)
	data = append(data, cmdUpdateDevice...)

	data = append(data, cmdLatticeStart...)
	data = append(data, commandRetractPaper(feed)...)
	for _, row := range img {
		data = append(data, commandPrintRow(row)...)
	}
	data = append(data, cmdLatticeEnd...)
	data = append(data, cmdFinalSpeed...)
	data = append(data, commandFeedPaper(feed)...)
	data = append(data, cmdSetPaper...)
	data = append(data, cmdSetPaper...)
	data = append(data, cmdSetPaper...)
	data = append(data, cmdGetDevState...)
	return data

}

func weakCommandsPrintImg(imgS []byte, feed int) []byte {

	img := chunkify(imgS, printWidth)

	data := append(cmdGetDevState, cmdSetQuality200Dpi...)
	data = append(data, cmdLatticeStart...)
	data = append(data, commandRetractPaper(feed)...)
	for _, row := range img {
		data = append(data, commandPrintRow(row)...)
	}
	data = append(data, commandFeedPaper(feed)...)
	data = append(data, cmdSetPaper...)
	data = append(data, cmdSetPaper...)
	data = append(data, cmdSetPaper...)
	data = append(data, cmdLatticeEnd...)
	data = append(data, cmdGetDevState...)
	return data

}
