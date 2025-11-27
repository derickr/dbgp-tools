package main

import (
	"context"
	"fmt"
	"net"
	"unsafe"

	"github.com/Microsoft/go-winio"
	"github.com/hillu/go-ntdll"
)

func findFiles() (map[int]string, error) {
	//var retval map[int]string
	retval := make(map[int]string)

	var h ntdll.Handle
	var oa = ntdll.ObjectAttributes{
		ObjectName: ntdll.NewUnicodeString(`\Device\NamedPipe\`),
		Attributes: 0x40, // OBJ_CASE_INSENSITIVA
	}
	oa.Length = uint32(unsafe.Sizeof(oa))

	var iosb ntdll.IoStatusBlock

	if st := ntdll.NtOpenFile(&h,
		ntdll.SYNCHRONIZE|ntdll.AccessMask(1), // FILE_LIST_DIRECTORY
		&oa,
		&iosb,
		ntdll.FILE_SHARE_READ|ntdll.FILE_SHARE_WRITE|ntdll.FILE_SHARE_DELETE,
		ntdll.FILE_DIRECTORY_FILE|ntdll.FILE_SYNCHRONOUS_IO_NONALERT|ntdll.FILE_OPEN_FOR_BACKUP_INTENT,
	); st.Error() != nil {
		// print error,
		return retval, st.Error()
	}

	buf := make([]byte, 64*1024) // thats a lot?
	restart := true

	for {
		var iosb ntdll.IoStatusBlock
		st := ntdll.NtQueryDirectoryFile(h,
			0,
			nil,
			nil,
			&iosb,
			&buf[0],
			uint32(len(buf)),
			ntdll.FileDirectoryInformation,
			false,
			ntdll.NewUnicodeString("xdebug-ctrl.*"),
			restart,
		)
		restart = false

		if !st.IsSuccess() {
			if st == ntdll.STATUS_NO_MORE_FILES {
				break
			}
			// print error
			return retval, st.Error()
		}

		for offset := 0; offset < int(iosb.Information); {

			fdi := (*ntdll.FileDirectoryInformationT)(unsafe.Pointer(&buf[offset]))

			fn := fdi.FileNameSlice(int(fdi.FileNameLength))
			filename := ntdll.NewUnicodeStringFromBuffer(&fn[0], int(fdi.FileNameLength))
			pid := 0
			if n, err := fmt.Sscanf(filename.String(), "xdebug-ctrl.%d", &pid); err == nil && n == 1 {
				retval[pid] = `\\.\pipe\` + filename.String()
			}

			if fdi.NextEntryOffset == 0 {
				break
			}
			offset += int(fdi.NextEntryOffset)
		}
	}

	return retval, nil
}

func dialCtrlSocket(ctx context.Context, ctrl_socket string) (net.Conn, error) {
	conn, err := winio.DialPipeContext(ctx, ctrl_socket)
	return conn, err
}
