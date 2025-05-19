package main

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var (
	clog_output_mutex = sync.Mutex{}
)

func clog_print(args ...interface{}) {
	clog_output_mutex.Lock()
	defer clog_output_mutex.Unlock()
	fmt.Print(args...)
}

func clog_println(args ...interface{}) {
	clog_output_mutex.Lock()
	defer clog_output_mutex.Unlock()
	fmt.Println(args...)
}

func clog_printf(format string, args ...interface{}) {
	clog_output_mutex.Lock()
	defer clog_output_mutex.Unlock()
	fmt.Printf(format, args...)
}

func custom_log(level string, format string, args ...interface{}) {
	if strings.EqualFold(level, "Debug") && !(clog_enable_debug || clog_enable_trace) {
		return
	}
	if strings.EqualFold(level, "Trace") && !clog_enable_trace {
		return
	}

	clog_output_mutex.Lock()
	defer clog_output_mutex.Unlock()

	var func_name string = ""
	var file string = ""
	var line int = 0

	if clog_enable_debug || clog_enable_trace {
		var pc uintptr
		var ok bool
		func_name = "?"
		pc, file, line, ok = runtime.Caller(1)
		if ok {
			func_obj := runtime.FuncForPC(pc)
			if func_obj != nil {
				func_name = func_obj.Name()
			}
		} else {
			file = "?"
			line = 0
		}
		file = path.Base(file)
		file = file + ":"
		func_name = "(" + func_name + ")"
	}

	if clog_logfile_handle != nil {
		clog_logfile_handle.WriteString(fmt.Sprintf("%s %s%d%s [%s] ", time.Now().Format("15:04:05"), file, line, func_name, level))
		clog_logfile_handle.WriteString(fmt.Sprintf(format, args...))
		clog_logfile_handle.WriteString("\n")
	}

	if clog_log_trace_to_file_only && strings.EqualFold(level, "Trace") {
		return
	}

	if !clog_enable_colored_output {
		// uncolored output
		fmt.Printf("%s %s%d%s [%s] ", time.Now().Format("15:04:05"), file, line, func_name, level)
		fmt.Printf(format, args...)
		fmt.Printf("\n")
	} else {
		// colored output
		var level_color *color.Color
		switch strings.ToLower(level) {
		case "trace":
			level_color = color.New(color.Reset, color.Faint, color.FgHiWhite, color.BgHiBlack)
		case "debug":
			level_color = color.New(color.Reset, color.FgBlack, color.BgYellow)
		case "info":
			level_color = color.New(color.Reset, color.FgBlack, color.BgHiWhite)
		case "warn":
			level_color = color.New(color.Reset, color.Bold, color.FgBlack, color.BgHiYellow)
		case "error":
			level_color = color.New(color.Reset, color.Bold, color.FgBlack, color.BgHiRed)
		case "fatal":
			level_color = color.New(color.Reset, color.Bold, color.FgBlack, color.BgHiMagenta)
		default:
			level_color = color.New(color.Reset, color.Bold, color.FgBlack, color.BgHiMagenta)
		}

		fmt.Printf("%s %s%d%s ", time.Now().Format("15:04:05.000"), file, line, func_name)
		level_color.Printf("[%s]", level)
		fmt.Printf(" ")
		fmt.Printf(format, args...)
		fmt.Printf("\n")
	}

}
