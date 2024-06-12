package processor

import (
	"encoding/binary"
	"fmt"
	"github.com/ngyewch/java-bytecode-version-scanner/scanner"
	"github.com/spf13/afero"
	"sort"
	"strings"
)

type Processor struct {
	scanContextMap map[*scanner.ScanContext]*Collector
}

type Collector struct {
	Entries []Entry
}

type Entry struct {
	Path    string
	Version ByteCodeVersion
}

type ByteCodeVersion struct {
	Major uint16
	Minor uint16
}

func (v ByteCodeVersion) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

func (v ByteCodeVersion) JavaVersion() string {
	if v.Major < 45 {
		return "unknown"
	} else if v.Major > 48 {
		return fmt.Sprintf("%d", v.Major-44)
	} else {
		return fmt.Sprintf("1.%d", v.Major-44)
	}
}

func NewProcessor() *Processor {
	return &Processor{
		scanContextMap: make(map[*scanner.ScanContext]*Collector),
	}
}

func (processor *Processor) Process(sc *scanner.ScanContext, path string) error {
	if !strings.HasSuffix(path, ".class") {
		return nil
	}
	if path == "module-info.class" {
		return nil
	}
	if strings.HasPrefix(path, "META-INF/versions/") {
		return nil
	}

	f, err := sc.FS.Open(path)
	if err != nil {
		return err
	}
	defer func(f afero.File) {
		_ = f.Close()
	}(f)

	var magic uint32
	var minorVersion uint16
	var majorVersion uint16
	err = binary.Read(f, binary.BigEndian, &magic)
	if err != nil {
		return err
	}
	err = binary.Read(f, binary.BigEndian, &minorVersion)
	if err != nil {
		return err
	}
	err = binary.Read(f, binary.BigEndian, &majorVersion)
	if err != nil {
		return err
	}

	if magic != 0xCAFEBABE {
		return nil
	}

	version := ByteCodeVersion{
		Major: majorVersion,
		Minor: minorVersion,
	}

	collector, ok := processor.scanContextMap[sc]
	if !ok {
		collector = &Collector{}
		processor.scanContextMap[sc] = collector
	}

	collector.Entries = append(collector.Entries, Entry{
		Path:    path,
		Version: version,
	})

	return nil
}

func (processor *Processor) Report(maxByteCodeMajorVersion uint16, listClassFiles bool) {
	for scanContext, collector := range processor.scanContextMap {
		collector.Report(scanContext, maxByteCodeMajorVersion, listClassFiles)
	}
}

type Bucket struct {
	Version ByteCodeVersion
	Entries []Entry
}

type BucketsByCount []*Bucket

func (a BucketsByCount) Len() int {
	return len(a)
}

func (a BucketsByCount) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a BucketsByCount) Less(i, j int) bool {
	return len(a[i].Entries) > len(a[j].Entries)
}

func (collector *Collector) Report(sc *scanner.ScanContext, maxByteCodeMajorVersion uint16, listClassFiles bool) {
	bucketMap := make(map[ByteCodeVersion]*Bucket)
	for _, entry := range collector.Entries {
		bucket, ok := bucketMap[entry.Version]
		if !ok {
			bucket = &Bucket{
				Version: entry.Version,
			}
			bucketMap[entry.Version] = bucket
		}
		bucket.Entries = append(bucket.Entries, entry)
	}
	var buckets []*Bucket
	for _, bucket := range bucketMap {
		buckets = append(buckets, bucket)
	}
	sort.Sort(BucketsByCount(buckets))
	if len(buckets) == 0 {
		return
	}
	hasProblems := false
	for _, bucket := range buckets {
		if bucket.Version.Major > maxByteCodeMajorVersion {
			hasProblems = true
			break
		}
	}
	if !hasProblems {
		return
	}
	fmt.Println()
	fmt.Println(strings.Join(sc.PathComponents(), " | "))
	for _, bucket := range buckets {
		fmt.Printf("%s (%s) - %d\n", bucket.Version, bucket.Version.JavaVersion(), len(bucket.Entries))
		if listClassFiles && (bucket.Version.Major > maxByteCodeMajorVersion) {
			for _, entry := range bucket.Entries {
				fmt.Printf("- %s\n", entry.Path)
			}
		}
	}
}
