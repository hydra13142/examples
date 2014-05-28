package main

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

func Display(w io.Writer, v interface{}) {
	var (
		name func(reflect.Type) string
		show func(int, reflect.Value, string)
	)
	name = func(t reflect.Type) string {
		x := t.Name()
		if x == "" {
			switch t.Kind() {
			case reflect.Interface:
				return "interface{}"
			case reflect.Struct:
				return "struct"
			case reflect.Ptr:
				return fmt.Sprintf("(*%s)", name(t.Elem()))
			case reflect.Array:
				return fmt.Sprintf("[...]%s", name(t.Elem()))
			case reflect.Slice:
				return fmt.Sprintf("[]%s", name(t.Elem()))
			case reflect.Map:
				return fmt.Sprintf("map[%s]%s", name(t.Key()), name(t.Elem()))
			}
			return x
		}
		pre := strings.Replace(t.PkgPath(), "/", ".", -1)
		if pre != "" {
			return pre + "." + x
		} else {
			return x
		}
	}
	show = func(a int, x reflect.Value, sf string) {
		if a >= 0 {
			for i := 0; i < a; i++ {
				fmt.Fprint(w, "    ")
			}
		} else {
			a = -a
		}
		defer fmt.Fprint(w, sf)

		if !x.IsValid() {
			fmt.Fprint(w, "nil")
			return
		}
		y := x.Type()
		switch y.Kind() {
		case reflect.Ptr:
			if !x.Elem().IsValid() {
				fmt.Fprint(w, "nil")
				return
			}
			fmt.Fprintf(w, "&")
			show(-a, x.Elem(), "")
		case reflect.Interface:
			fmt.Fprintf(w, "%s(\r\n", name(y))
			show(a+1, x.Elem(), "\r\n")
			for i := 0; i < a; i++ {
				fmt.Fprint(w, "    ")
			}
			fmt.Fprint(w, ")")
		case reflect.Struct:
			fmt.Fprintf(w, "%s{", name(y))
			if x.NumField() != 0 {
				fmt.Fprint(w, "\r\n")
				for i := 0; i < x.NumField(); i++ {
					for i := 0; i <= a; i++ {
						fmt.Fprint(w, "    ")
					}
					fmt.Fprintf(w, "%s:", y.Field(i).Name)
					show(-a-1, x.Field(i), ",\r\n")
				}
				for i := 0; i < a; i++ {
					fmt.Fprint(w, "    ")
				}
			}
			fmt.Fprint(w, "}")
		case reflect.Slice, reflect.Array:
			fmt.Fprintf(w, "%s{", name(y))
			if x.Len() != 0 {
				fmt.Fprint(w, "\r\n")
				for i := 0; i < x.Len(); i++ {
					show(a+1, x.Index(i), ",\r\n")
				}
				for i := 0; i < a; i++ {
					fmt.Fprint(w, "    ")
				}
			}
			fmt.Fprint(w, "}")
		case reflect.Map:
			fmt.Fprintf(w, "%s{", name(y))
			if x.Len() != 0 {
				fmt.Fprint(w, "\r\n")
				for _, k := range x.MapKeys() {
					show(a+1, k, ":")
					show(-a-1, x.MapIndex(k), ",\r\n")
				}
				for i := 0; i < a; i++ {
					fmt.Fprint(w, "    ")
				}
			}
			fmt.Fprint(w, "}")
		case reflect.Bool:
			fmt.Fprintf(w, "%t", x.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fmt.Fprintf(w, "%d", x.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fmt.Fprintf(w, "%d", x.Uint())
		case reflect.Float32, reflect.Float64:
			fmt.Fprintf(w, "%g", x.Float())
		case reflect.Complex64, reflect.Complex128:
			fmt.Fprintf(w, "%v", x.Complex())
		case reflect.Uintptr:
			fmt.Fprintf(w, "%p", x.Uint())
		case reflect.String:
			fmt.Fprintf(w, `"%s"`, x.String())
		default:
			fmt.Fprintf(w, "%s", x.String())
		}
	}
	show(0, reflect.ValueOf(v), "\r\n")
}

func main() {
	Display(os.Stdout, map[string]string{"x": "y"})
}
