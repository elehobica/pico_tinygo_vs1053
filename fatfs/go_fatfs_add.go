package fatfs

// #include <string.h>
// #include <stdlib.h>
// #include "./go_fatfs.h"
import "C"

func (l *FATFS) GetFsType() (Type, error) {
	return Type(l.fs.fs_type), nil
}

func (l *FATFS) GetCardSize() (int64, error) {
	return int64(l.fs.csize) * int64(l.fs.n_fatent) * SectorSize, nil
}

// Seek changes the position of the file
func (f *File) Seek(offset int64) error {
	var ofs C.FSIZE_t = C.FSIZE_t(offset)
	errno := C.f_lseek(f.fileptr(), ofs)
	return errval(errno)
}

func (f *File) Tell() (ret int64, err error) {
	pos := int64(f.fileptr().fptr)
	if pos < 0 {
		return -1, errval(C.FRESULT(C.FR_INT_ERR))
	}
	return int64(pos), nil
}

// Rewind changes the position of the file to the beginning of the file
func (f *File) Rewind() (err error) {
	return f.Seek(0)
}

// Truncates the size of the file to the specified size
//
func (f *File) Truncate() error {
	return errval(C.f_truncate(f.fileptr()))
}

//  Allocate a contiguous block to the file
//
func (f *File) Expand(size int64, flag bool) error {
	var fsz C.FSIZE_t = C.FSIZE_t(size)
	var opt C.BYTE;
	if flag { opt = 1 } else { opt = 0 }
	return errval(C.f_expand(f.fileptr(), fsz, opt))
}