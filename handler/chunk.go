package handler

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"sort"
)

const (
	maxChunkSize = int64(5 << 20) // 5MB
	uploadDir    = "./data/chunks"
)

type Chunk struct {
	UploadID      string // unique id for the current upload.
	ChunkNumber   int32
	TotalChunks   int32
	TotalFileSize int64 // in bytes
	Filename      string
	Data          io.Reader
	UploadDir     string
}

func ProcessChunk(r *http.Request) error {
	chunk, err := ParseChunk(r)
	if err != nil {
		return fmt.Errorf("failed to parse chunk %w", err)
	}

	if err := os.MkdirAll(chunk.UploadID, 02750); err != nil {
		return err
	}

	if err := StoreChunk(chunk); err != nil {
		return err
	}

	return nil
}

func CompleteChunk(uploadID, filename string) error {
	uploadDir := fmt.Sprintf("%s/%s", uploadDir, uploadID)

	err := RebuildFile(uploadDir, filename)
	if err != nil {
		return fmt.Errorf("failed to rebuild file %w", err)
	}

	return nil
}

func ParseChunk(r *http.Request) (*Chunk, error) {
	var chunk Chunk

	buf := new(bytes.Buffer)

	reader, err := r.MultipartReader()
	if err != nil {
		return nil, err
	}

	// start readings parts
	// 1. upload id
	// 2. chunk number
	// 3. total chunks
	// 4. total file size
	// 5. file name
	// 6. chunk data

	// 1
	if err := getPart("id", reader, buf); err != nil {
		return nil, err
	}

	chunk.UploadID = buf.String()
	buf.Reset()

	// dir to where we store our chunk
	chunk.UploadDir = fmt.Sprintf("%s/%s", uploadDir, chunk.UploadID)

	// 2
	if err := getPart("num", reader, buf); err != nil {
		return nil, err
	}

	parsedChunkNumber, err := strconv.ParseInt(buf.String(), 10, 32)
	if err != nil {
		return nil, err
	}

	chunk.ChunkNumber = int32(parsedChunkNumber)
	buf.Reset()

	// 3
	if err := getPart("total", reader, buf); err != nil {
		return nil, err
	}

	parsedTotalChunksNumber, err := strconv.ParseInt(buf.String(), 10, 32)
	if err != nil {
		return nil, err
	}

	chunk.TotalChunks = int32(parsedTotalChunksNumber)
	buf.Reset()

	// 4
	if err := getPart("size", reader, buf); err != nil {
		return nil, err
	}

	parsedTotalFileSizeNumber, err := strconv.ParseInt(buf.String(), 10, 64)
	if err != nil {
		return nil, err
	}

	chunk.TotalFileSize = parsedTotalFileSizeNumber
	buf.Reset()

	// 5
	if err := getPart("name", reader, buf); err != nil {
		return nil, err
	}

	chunk.Filename = buf.String()
	buf.Reset()

	// 6
	part, err := reader.NextPart()
	if err != nil {
		return nil, fmt.Errorf("failed reading chunk part %w", err)
	}

	chunk.Data = part

	return &chunk, nil
}

type ByChunk []os.FileInfo

func (a ByChunk) Len() int      { return len(a) }
func (a ByChunk) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByChunk) Less(i, j int) bool {
	ai, _ := strconv.Atoi(a[i].Name())
	aj, _ := strconv.Atoi(a[j].Name())
	return ai < aj
}

func StoreChunk(chunk *Chunk) error {
	chunkFile, err := os.Create(fmt.Sprintf("%s/%d", chunk.UploadDir, chunk.ChunkNumber))
	if err != nil {
		return err
	}

	if _, err := io.CopyN(chunkFile, chunk.Data, maxChunkSize); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func RebuildFile(dir string, name string) error {
	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	fullFile, err := os.Create(name)
	if err != nil {
		return err
	}
	sort.Sort(ByChunk(fileInfos))
	for _, fs := range fileInfos {
		if err := appendChunk(dir, fs, fullFile); err != nil {
			return err
		}
	}

	defer fullFile.Close()

	// if err := os.RemoveAll(uploadDir); err != nil {
	// 	return nil, err
	// }

	return nil
}

func appendChunk(uploadDir string, fs os.FileInfo, fullFile *os.File) error {
	src, err := os.Open(uploadDir + "/" + fs.Name())

	if err != nil {
		return err
	}
	defer src.Close()
	if _, err := io.Copy(fullFile, src); err != nil {
		return err
	}

	return nil
}

func getPart(expectedPart string, reader *multipart.Reader, buf *bytes.Buffer) error {
	part, err := reader.NextPart()
	if err != nil {
		return fmt.Errorf("failed reading %s part %w", expectedPart, err)
	}

	if part.FormName() != expectedPart {
		return fmt.Errorf("invalid form name for part. Expected %s got %s", expectedPart, part.FormName())
	}

	if _, err := io.Copy(buf, part); err != nil {
		return fmt.Errorf("failed copying %s part %w", expectedPart, err)
	}

	return nil
}
