//package logtail
package main

import (
	"log"
	"bufio"
	"os"
	"flag"
	"time"
	"fmt"
	"strconv"
	"strings"
)

func parse_range(r string) (int, int, error) {
	
	parts := strings.Split(r, ":")
	var min, max int
	var err error
	
	min, max = -1, -1
	
	if len(parts) == 1 {
		max, err = strconv.Atoi(r)
		if err != nil {
			return -1, -1, fmt.Errorf("invalid max value: %s", r)
		}
	} else {
		if parts[0] != "" {
			min, err = strconv.Atoi(parts[0])
			if err != nil {
				return -1, -1, fmt.Errorf("invalid min value: %s", parts[0])
			}
		}
		if parts[1] != "" {
			max, err = strconv.Atoi(parts[1])
			if err != nil {
				return -1, -1, fmt.Errorf("invalid max value: %s", parts[1])
			}
		}
	}
	
	return min, max, nil
}

func main() {
	
	var prefix *string
	max_memory_bytes := -1
	min_bytes := -1
	max_bytes := -1
	min_lines := -1
	max_lines := -1
	out_file := "stdout.log"
	rfc_3339 := false
	
	flag.Func("p", "Prefix a line with a given string", func(val string) error {
		prefix = &val
		return nil
	})
	flag.IntVar(&max_memory_bytes, "m", -1, "Maximum in-memory buffer size in bytes")
	
	var bytes_range, lines_range string
	flag.StringVar(&bytes_range, "c", "", "Byte range in the format [min:]max")
	flag.StringVar(&lines_range, "n", "", "Line range in the format [min:]max")
	
	flag.BoolVar(&rfc_3339, "d", false, "Print date/time in RFC 3339 format before each line/prefix")
	
	flag.Parse()
	
	if bytes_range != "" {
		var err error
		min_bytes, max_bytes, err = parse_range(bytes_range)
		if err != nil {
			fmt.Println("error: byte range parse error (option -c):", err)
			os.Exit(1)
		}
	} else {
		min_bytes, max_bytes = -1, -1
	}
	if lines_range != "" {
		var err error
		min_lines, max_lines, err = parse_range(lines_range)
		if err != nil {
			fmt.Println("error: line range parse error (option -n):", err)
			os.Exit(1)
		}
	} else {
		min_lines, max_lines = -1, -1
	}
	
	args := flag.Args()
	if len(args) > 0 {
		out_file = args[0]
	}
	
	if min_bytes == -1 {
		min_bytes = max_bytes
	}
	if min_lines == -1 {
		min_lines = max_lines
	}
	
	// parse args and update options
	bytes_per_line := []int{}
	out_lines := 0
	out_bytes := 0
	
	// read number of out_lines and out_bytes from out_file
	fh, err := os.OpenFile(out_file, os.O_RDWR | os.O_CREATE, 0644)
	if err != nil {
		os.Exit(1)
	}
	defer fh.Close()
	
	autotrim := func(n_bytes int) {
		trim_bytes := 0
		trim_lines := 0
		
		// append to out_file if out_lines/out_bytes is within maximum range
		if max_lines != -1 && out_lines+1 > max_lines {
			// we MUST trim lines to fit
			
			// log.Printf("info: max lines reached: %d + %d > %d\n", out_lines, 1, max_lines)
			
			n := len(bytes_per_line)
			for i := 0; i < n && out_lines+1 - trim_lines > min_lines; i += 1 {
				trim_lines += 1
				trim_bytes += bytes_per_line[i]
			}
		}
		if max_bytes != -1 && out_bytes+n_bytes - trim_bytes > max_bytes {
			// we MUST trim bytes to fit
			
			// log.Printf("info: max bytes reached: %d + %d > %d\n", out_bytes, n_bytes, max_bytes)
			
			n := len(bytes_per_line)
			for i := 0; i < n && out_bytes+n_bytes - trim_bytes > min_bytes; i += 1 {
				trim_bytes += bytes_per_line[i]
				trim_lines += 1
			}
		}
		
		if trim_bytes > 0 {
			// go to trim_bytes from start of file (use chunks of max_memory_bytes, or the entire file)
			
			// log.Printf("info: trimming: %d lines, %d bytes\n", trim_lines, trim_bytes)
			
			write_bytes := out_bytes - trim_bytes
			
			write_offset := 0
			read_offset := trim_bytes
			
			bufsize := max_memory_bytes
			if bufsize == -1 {
				bufsize = write_bytes
			}
			
			buffer := make([]byte, bufsize)
			for write_offset < write_bytes {
				
				// go to the start of read
				fh.Seek(int64(read_offset), 0)
				// read buffer
				n, _ := fh.Read(buffer)
				// go to the start of write
				fh.Seek(int64(write_offset), 0)
				// write buffer
				fh.Write(buffer[:n])
				
				write_offset += n
				read_offset += n
			}
			
			// now truncate the file
			fh.Truncate(int64(write_bytes))
			
			// go to the end of the file
			fh.Seek(0, 2)
			
			out_bytes -= trim_bytes
			out_lines -= trim_lines
			
			bytes_per_line = bytes_per_line[trim_lines:]
			
			// log.Printf("info: trimmed to: %d lines, %d bytes\n", out_lines, out_bytes)
		}
	}
	
	// go to the start of the file (don't append at the back)
	fh.Seek(0, 0)
	
	// read lines from existing file, and update variables:
	s := bufio.NewScanner(fh)
	for s.Scan() {
		line := s.Text()
		n := len(line) + 1
		out_lines += 1
		out_bytes += n
		bytes_per_line = append(bytes_per_line, n)
	}
	if err := s.Err(); err != nil {
		log.Println(err)
	}
	
	// log.Printf("info: read existing: %d lines, %d bytes\n", out_lines, out_bytes)
	
	// trim existing file
	autotrim(0)
	
	// now read from stdin and append stdin to file, but keep trimming when needed (check every line)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		
		line := ""
		if rfc_3339 {
			// note: RFC3339 is a stricter version of ISO8601
			line += time.Now().Format(time.RFC3339)
			
			// if no prefix given, automatically inject a tab
			if prefix == nil {
				line += "\t"
			} else {
				line += *prefix
			}
		} else if prefix != nil {
			line += *prefix
		}
		line += scanner.Text() + "\n"
		n := len(line)
		
		autotrim(n)
		
		fh.WriteString(line)
		out_lines += 1
		out_bytes += n
		bytes_per_line = append(bytes_per_line, n)
	}
	
	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}
