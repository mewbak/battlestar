package main

// TODO: Add line numbers to the error messages and make them parseable by editors and IDEs

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type (
	Program   string
	TokenType int
	Token     struct {
		t     TokenType
		value string
	}
	TokenDescriptions map[TokenType]string
	Statement         []Token
)

const (
	REGISTER   = 0
	ASSIGNMENT = 1
	VALUE      = 2
	KEYWORD    = 3
	BUILTIN    = 4
	VALID_NAME = 5
	STRING     = 6
	DISREGARD  = 7
	RESERVED   = 8
	VARIABLE   = 9
	SEP        = 127
	UNKNOWN    = 255
)

// Global variables
var (
	in_function   string   // name of the function we are currently in
	defined_names []string // all defined variables/constants/functions
	variables map[string][]string // list of variable names per function name

	registers = []string{"ah", "bh", "ch", "dh", "si", "di", "sp", "bp", "ip", // 16-bit
		"eax", "ebx", "ecx", "edx", "esi", "edi", "esp", "ebp", "eip", // 32-bit
		"rax", "rbx", "rcx", "rdx", "rsi", "rdi", "rsp", "rbp", "rip", "r8", "r9", "r10", "r11", "r12", "r13", "r14", "r15", "sil", "dil", "spl", "bpl", "xmm8", "xmm9", "xmm10", "xmm11", "xmm12", "xmm13", "xmm14", "xmm15"} // 64-bit

	operators = []string{"=", "+=", "-=", "*=", "/="}
	keywords  = []string{"fun", "ret", "const", "call", "extern", "end", "var"}
	builtins  = []string{"len", "int", "exit"} // built-in functions
	reserved  = []string{"param"}              // built-in variables

	token_to_string = TokenDescriptions{REGISTER: "register", ASSIGNMENT: "assignment", VALUE: "value", VALID_NAME: "name", SEP: ";", UNKNOWN: "?", KEYWORD: "keyword", STRING: "string", BUILTIN: "built-in", DISREGARD: "disregard", RESERVED: "reserved", VARIABLE: "variable"}

	// 32-bit (i686) or 64-bit (x86_64)
	platform_bits = 32

	// OS X or Linux
	osx = false

	// TODO: Add an option for not adding start symbols
	linker_start_function = "_start"

	// TODO: Add an option for not adding an exit function
)

// Check if a given map has a given key
func haskey(sm map[TokenType]string, key TokenType) bool {
	_, present := sm[key]
	return present
}

func (tok Token) String() string {
	if tok.t == SEP {
		return ";"
	} else if haskey(token_to_string, tok.t) {
		return token_to_string[tok.t] + ":" + tok.value
	}
	log.Fatalln("Error when serializing: Unfamiliar token type when representing token as string: " + tok.value)
	return "!?"
}

// Maps the function f over a slice of strings
func maps(sl []string, f func(string) string) []string {
	newl := make([]string, len(sl), len(sl))
	for i, element := range sl {
		newl[i] = f(element)
	}
	return newl
}

// Checks if a slice of strings has the given string
func has(sl []string, s string) bool {
	for _, e := range sl {
		if e == s {
			return true
		}
	}
	return false
}

func is_valid_name(s string) bool {
	if len(s) == 0 {
		return false
	}
	// TODO: These could be global constants instead
	letters := "abcdefghijklmnopqrstuvwxyz"
	upper := strings.ToUpper(letters)
	digits := "0123456789"
	special := "_·"
	combined := letters + upper + digits + special

	// Does not start with a number
	if strings.Contains(digits, string(s[0])) {
		return false
	}
	// Check that the rest are valid characters
	for _, letter := range s {
		// If not a letter, digit or valid special character, it's not a valid name
		if !(strings.Contains(combined, string(letter))) {
			return false
		}
	}
	// Valid
	return true
}

// Split a string into more tokens and tokenize them
func retokenize(word string, sep string, debug bool) []Token {
	var newtokens []Token
	words := strings.Split(word, sep)
	for _, s := range words {
		tokens := tokenize(s, debug)
		//log.Println("RETOKEN", tokens)
		for _, t := range tokens {
			if t.t != SEP {
				newtokens = append(newtokens, t)
			}
		}
	}
	return newtokens
}

// Remove one line commants, both // and # are ok
func removecomments(s string) string {
	if strings.HasPrefix(s, "//") || strings.HasPrefix(s, "#") {
		return ""
	}
	return s
}

// Tokenize a string
func tokenize(program string, debug bool) []Token {
	statements := maps(maps(strings.Split(program, "\n"), strings.TrimSpace), removecomments)
	tokens := make([]Token, 0, 0)
	var t Token
	var instring bool    // Have we encountered a " for any given statement?
	var collected string // Collected string, until end of line
	for _, statement := range statements {
		words := maps(strings.Split(statement, " "), strings.TrimSpace)
		for _, word := range words {
			if word == "" {
				continue
			}
			// TODO: refactor out code that repeats the same thing
			if instring {
				collected += word + " "
			} else if has(registers, word) {
				t = Token{REGISTER, word}
				tokens = append(tokens, t)
				if debug {
					log.Println("TOKEN", t)
				}
			} else if has(operators, word) {
				t = Token{ASSIGNMENT, word}
				tokens = append(tokens, t)
				if debug {
					log.Println("TOKEN", t)
				}
			} else if has(keywords, word) {
				t = Token{KEYWORD, word}
				tokens = append(tokens, t)
				if debug {
					log.Println("TOKEN", t)
				}
			} else if has(builtins, word) {
				t = Token{BUILTIN, word}
				tokens = append(tokens, t)
				if debug {
					log.Println("TOKEN", t)
				}
			} else if has(reserved, word) {
				t = Token{RESERVED, word}
				tokens = append(tokens, t)
				if debug {
					log.Println("TOKEN", t)
				}
			} else if _, err := strconv.Atoi(word); err == nil {
				t = Token{VALUE, word}
				tokens = append(tokens, t)
				if debug {
					log.Println("TOKEN", t)
				}
			} else if word == "_" {
				t = Token{DISREGARD, word}
				tokens = append(tokens, t)
				if debug {
					log.Println("TOKEN", t)
				}
			} else if is_valid_name(word) {
				t = Token{VALID_NAME, word}
				tokens = append(tokens, t)
				if debug {
					log.Println("TOKEN", t)
				}
			} else if strings.Contains(word, "(") {
				if debug {
					log.Println("RETOKENIZE BECAUSE OF \"(\"")
				}
				newtokens := retokenize(word, "(", debug)
				for _, newtoken := range newtokens {
					tokens = append(tokens, newtoken)
				}
				log.Println("NEWTOKENS", newtokens)
			} else if strings.Contains(word, ")") {
				if debug {
					log.Println("RETOKENIZE BECAUSE OF \")\"")
				}
				newtokens := retokenize(word, ")", debug)
				for _, newtoken := range newtokens {
					tokens = append(tokens, newtoken)
				}
				log.Println("NEWTOKENS", newtokens)
			} else if strings.Contains(word, "[") {
				if debug {
					log.Println("RETOKENIZE BECAUSE OF \"[\"")
				}
				newtokens := retokenize(word, "[", debug)
				for _, newtoken := range newtokens {
					tokens = append(tokens, newtoken)
				}
				log.Println("NEWTOKENS", newtokens)
			} else if strings.Contains(word, "]") {
				if debug {
					log.Println("RETOKENIZE BECAUSE OF \"]\"")
				}
				newtokens := retokenize(word, "]", debug)
				for _, newtoken := range newtokens {
					tokens = append(tokens, newtoken)
				}
				log.Println("NEWTOKENS", newtokens)
			} else if strings.Contains(word, ",") {
				if debug {
					log.Println("RETOKENIZE BECAUSE OF \",\"")
				}
				newtokens := retokenize(word, ",", debug)
				for _, newtoken := range newtokens {
					tokens = append(tokens, newtoken)
				}
				log.Println("NEWTOKENS", newtokens)
			} else if strings.Contains(word, "\"") {
				if debug {
					log.Println("TOKEN", word, "is part of a string")
					log.Println("ENTERING STRING")
				}
				instring = true
				collected = word + " "
			} else {
				if debug {
					log.Println("TOKEN", word, "unknown")
				}
				log.Fatalln("Error: Unrecognized token:", word)
				return tokens
			}
		}
		if instring {
			if debug {
				log.Println("EXITING STRING AT END OF STATEMENT")
				log.Println("STRING:", collected)
				t = Token{STRING, collected}
				tokens = append(tokens, t)
			}
			instring = false
		}
		t = Token{SEP, ";"}
		tokens = append(tokens, t)
	}
	return tokens
}

// Replace built-in function calls with more basic code
func reduce(st Statement, debug bool) Statement {
	for i := 0; i < (len(st) - 1); i++ {
		// The built-in len() function
		if (st[i].t == BUILTIN) && (st[i].value == "len") && (st[i+1].t == VALID_NAME) {
			// len followed by a valid name
			// replace with the length of the given value

			name := st[i+1].value

			if !has(defined_names, name) {
				log.Fatalln("Error:", name, "is unfamiliar. Can not find length.")
			}
			if has(variables[in_function], name) {
				// TODO: Find a way to find the length of local variables
				log.Fatalln("Error: finding the length of a local variable is currently not implemented")
			}

			token_type := st[i+1].t

			// remove the element at i+1
			st = st[:i+1+copy(st[i+1:], st[i+2:])]

			// replace len(name) with _length_of_name
			st[i] = Token{token_type, "_length_of_" + name}

			if debug {
				log.Println("SUCCESSFULL REPLACEMENT WITH", st[i])
			}
		}
	}
	return st
}

func (st Statement) String() string {
	debug := true

	reduced := reduce(st, debug)
	if len(reduced) != len(st) {
		return reduced.String()
	}
	if len(st) == 0 {
		log.Fatalln("Error: Empty statement.")
		return ""
	} else if (st[0].t == BUILTIN) && (st[0].value == "int") { // interrrupt call
		asmcode := "\t;--- call interrupt 0x" + st[1].value + " ---\n"
		// Check the number of parameters
		if len(st) > 6 {
			log.Println("Error: Too many parameters for interrupt call:")
			for _, t := range st {
				log.Println(t.value)
			}
			os.Exit(1)
		}
		// Store each of the parameters to the appropriate registers
		var (
			reg                           string
			n                             string
			comment                       string
			interrupt_parameter_registers []string
		)
		if platform_bits == 32 {
			interrupt_parameter_registers = []string{"eax", "ebx", "ecx", "edx"}
		} else {
			interrupt_parameter_registers = []string{"rax", "rbx", "rcx", "rdx"}
		}
		for i := 2; i < len(st); i++ {
			reg = interrupt_parameter_registers[i-2]
			n = strconv.Itoa(i - 2)
			if (i - 2) == 0 {
				comment = "function call: " + st[i].value
			} else {
				if st[i].t == VALUE {
					comment = "parameter #" + n + " is " + st[i].value
				} else if st[i].t == REGISTER {
					log.Fatalln("Error: Can't use a register as a parameter to interrupt calls, since they may be overwritten when preparing for the call.\n" +
						"You can, however, use _ as a parameter to use the value in the corresponding register.")
				} else {
					if strings.HasPrefix(st[i].value, "_length_of_") {
						comment = "parameter #" + n + " is len(" + st[i].value[11:] + ")"
					} else {
						if st[i].value == "_" {
							// When _ is given, use the value already in the corresponding register
							comment = "parameter #" + n + " is already set"
						} else {
							comment = "parameter #" + n + " is " + "&" + st[i].value
						}
					}
				}
			}
			codeline := ""
			// Skip parameters/registers that are already set
			if st[i].value == "_" {
				codeline = "\t\t"
				log.Println("Note: Skipping the value for register " + reg + " when calling interrupt " + st[1].value)
			} else {
				codeline = "\tmov " + reg + ", " + st[i].value
			}

			// TODO: Find a more elegant way to format the comments in columns
			if len(codeline) > 14 { // for tab formatting
				asmcode += codeline + "\t\t; " + comment + "\n"
			} else {
				asmcode += codeline + "\t\t\t; " + comment + "\n"
			}
		}
		// Add the interrupt call
		if st[1].t == VALUE {
			asmcode += "\tint 0x" + st[1].value + "\t\t\t; perform the call\n"
			return asmcode
		}
		log.Fatalln("Error: Need a (hexadecimal) interrupt number to call:\n", st[1].value)
	} else if (st[0].t == KEYWORD) && (st[0].value == "const") && (len(st) >= 4) { // constant data
		constname := ""
		if st[1].t == VALID_NAME {
			constname = st[1].value
		} else {
			log.Fatalln(st[1].value, "is not a valid name for a constant")
		}
		asmcode := ""
		if (st[1].t == VALID_NAME) && (st[2].t == ASSIGNMENT) && ((st[3].t == STRING) || (st[3].t == VALUE) || (st[3].t == VALID_NAME)) {
			if has(defined_names, constname) {
				log.Fatalln("Error: Can not declare constant, name is already defined: " + constname)
			}
			if (st[3].t == VALID_NAME) && !has(defined_names, st[3].value) {
				log.Fatalln("Error: Can't assign", st[3].value, "to", st[1].value, "because", st[3].value, "is undefined.")
			}
			// Store the name of the declared constant in defined_names
			defined_names = append(defined_names, constname)
			// For the .DATA section (recognized by the keyword)
			asmcode += constname + ":\tdb "
			for i := 3; i < len(st); i++ {
				asmcode += st[i].value
				// Add a comma between every element but the last one
				if (i + 1) != len(st) {
					asmcode += ", "
				}
			}
			if st[3].t == STRING {
				asmcode += "\t\t; constant string\n"
			} else {
				asmcode += "\t\t; constant value\n"
			}
			// Special naming for storing the length for later
			asmcode += "_length_of_" + constname + " equ $ - " + constname + "\t; size of constant value\n"
			return asmcode
		}
		log.Println("Error: Invalid parameters for constant string statement:")
		for _, t := range st {
			log.Println(t.value)
		}
		os.Exit(1)
	} else if ((st[0].t == KEYWORD) && (st[0].value == "ret")) || ((st[0].t == BUILTIN) && (st[0].value == "exit")) {
		asmcode := ""
		if st[0].value == "ret" {
			if (in_function == "main") || (in_function == linker_start_function) {
				log.Println("Not taking down stack frame in the main/_start/start function.")
			} else {
				asmcode += "\t;--- takedown stack frame ---\n"
				if platform_bits == 32 {
					asmcode += "\tmov esp, ebp\t\t\t; use base pointer as new stack pointer\n"
					asmcode += "\tpop ebp\t\t\t\t; get the old base pointer\n\n"
				} else {
					asmcode += "\tmov rsp, rbp\t\t\t; use base pointer as new stack pointer\n"
					asmcode += "\tpop rbp\t\t\t\t; get the old base pointer\n\n"
				}
			}
		}
		if in_function != "" {
			asmcode += "\t;--- return from \"" + in_function + "\" ---\n"
		} else if st[0].value == "exit" {
			asmcode += "\t;--- exit program ---\n"
		} else {
			asmcode += "\t;--- return ---\n"
		}
		if (len(st) == 2) && (st[1].t == VALUE) {
			if platform_bits == 32 {
				asmcode += "\tmov eax, " + st[1].value + "\t\t\t; Error code "
			} else {
				asmcode += "\tmov rax, " + st[1].value + "\t\t\t; Error code "
			}
			if st[1].value == "0" {
				asmcode += "0 (everything is fine)\n"
			} else {
				asmcode += st[1].value + "\n"
			}
		}
		if (st[0].value == "exit") || (in_function == "main") || (in_function == linker_start_function) {
			// Not returning from main/_start/start function, but exiting properly
			exit_code := "0"
			if (len(st) == 2) && (st[1].t == VALUE) {
				exit_code = st[1].value
			}
			if platform_bits == 32 {
				asmcode += "\tmov eax, 1\t\t\t; function call: 1\n\tmov ebx, " + exit_code + "\t\t\t; return code " + exit_code + "\n\tint 0x80\t\t\t; exit program\n"
			} else {
				asmcode += "\tmov rax, 1\t\t\t; function call: 1\n\tmov rbx, " + exit_code + "\t\t\t; return code " + exit_code + "\n\tint 0x80\t\t\t; exit program\n"
			}
		} else {
			log.Println("IN FUNCTION", in_function)
			// Do not return eax=0/rax=0 if no return value is explicitly provided, by design
			// This allows the return value from the previous call to be returned instead
			asmcode += "\tret\t\t\t\t; Return\n"
		}
		if in_function != "" {
			// Exiting from the function definition
			in_function = ""
		}
		return asmcode
	} else if (st[0].t == REGISTER) || (st[0].t == DISREGARD) && (len(st) == 3) {
		// Statements like "eax = 3" are handled here
		// TODO: Handle all sorts of equivivalents to assembly statements
		if (st[0].t == REGISTER) && (st[1].t == ASSIGNMENT) && (st[2].t == VALUE || st[2].t == VALID_NAME) {
			return "\tmov " + st[0].value + ", " + st[2].value + "\t\t; " + st[0].value + " " + st[1].value + " " + st[2].value
		} else if (st[0].t == VALID_NAME) && (st[1].t == ASSIGNMENT) {
			if has(defined_names, st[0].value) {
				log.Fatalln("Error:", st[0].value, "has already been defined")
			} else {
				log.Fatalln("Error:", st[0].value, "is not recognized as a register (and there is no const qualifier). Can't assign.")
			}
		} else if (st[0].t == DISREGARD) && (st[1].t == ASSIGNMENT) {
			// TODO: If st[2] is a function, one wishes to call it, then disregard afterwards
			return "\t\t\t\t; Disregarding: " + st[2].value + "\n"
		} else if (st[0].t == REGISTER) && (st[1].t == ASSIGNMENT) && (st[2].t == REGISTER) {
			return "\tmov " + st[0].value + ", " + st[2].value + "\t\t\t; " + st[0].value + " " + st[1].value + " " + st[2].value
		} else if (st[0].t == RESERVED) && (st[1].t == VALUE) {
			if st[0].value == "param" {
				reg := "rbp"
				if platform_bits == 32 {
					reg = "ebp"
				}
				paramoffset, err := strconv.Atoi(st[1].value)
				if err != nil {
					log.Fatalln("Error: Invalid list offset for", st[0].value+":", st[1].value)
				}
				offset := strconv.Itoa(8 + paramoffset*4)
				return "[" + reg + "+" + offset + "]"
			} else {
				// TODO: Implement support for other lists
				log.Fatalln("Error: Can only handle \"param\" reserved words.")
			}
		} else if (st[0].t == REGISTER) && (st[1].t == ASSIGNMENT) && (st[2].t == RESERVED) && (st[3].t == VALUE) {
			if st[2].value == "param" {
				reg := "rbp"
				if platform_bits == 32 {
					reg = "ebp"
				}
				paramoffset, err := strconv.Atoi(st[3].value)
				if err != nil {
					log.Fatalln("Error: Invalid list offset for", st[2].value+":", st[3].value)
				}
				offset := strconv.Itoa(8 + paramoffset*4)
				return "\tmov " + st[0].value + ", " + "[" + reg + "+" + offset + "]\t\t; fetch function param #" + st[3].value + "\n"
			} else {
				// TODO: Implement support for other lists
				log.Fatalln("Error: Can only handle \"param\" lists.")
			}
		} else {
			log.Println("Error: Uknown type of 3 token statement:")
			for _, t := range st {
				log.Println("\t", t)
			}
			os.Exit(1)
		}
	} else if (len(st) >= 2) && (st[0].t == KEYWORD) && (st[1].t == VALID_NAME) && (st[0].value == "fun") {
		if in_function != "" {
			log.Fatalf("Error: Missing \"ret\"? Already in a function named %s when declaring function %s.\n", in_function, st[1].value)
		}
		asmcode := ";--- function " + st[1].value + " ---\n"
		in_function = st[1].value
		// Store the name of the declared function in defined_names
		if has(defined_names, in_function) {
			log.Fatalln("Error: Can not declare function, name is already defined:", in_function)
		}
		defined_names = append(defined_names, in_function)
		asmcode += "global " + in_function + "\t\t\t; make label available to the linker\n"
		asmcode += in_function + ":\t\t\t\t; name of the function\n\n"
		if (in_function == "main") || (in_function == linker_start_function) {
			log.Println("Note: Not setting up stack frame in the main/_start/start function.")
			return asmcode
		}
		asmcode += "\t;--- setup stack frame ---\n"
		if platform_bits == 32 {
			asmcode += "\tpush ebp\t\t\t; save old base pointer\n"
			asmcode += "\tmov ebp, esp\t\t\t; use stack pointer as new base pointer\n"
		} else {
			asmcode += "\tpush rbp\t\t\t; save old base pointer\n"
			asmcode += "\tmov rbp, rsp\t\t\t; use stack pointer as new base pointer\n"
		}
		return asmcode
	} else if (st[0].t == KEYWORD) && (st[0].value == "call") && (len(st) == 2) {
		if st[1].t == VALID_NAME {
			return "\t;--- call the \"" + st[1].value + "\" function ---\n\tcall " + st[1].value + "\n"
		} else {
			log.Fatalln("Calling an invalid name:", st[1].value)
		}
		// TODO: Find a shorter format to describe matching tokens.
		// Something along the lines of: if match(st, [KEYWORD:"extern"], 2)
	} else if (st[0].t == KEYWORD) && (st[0].value == "extern") && (len(st) == 2) {
		if st[1].t == VALID_NAME {
			extname := st[1].value
			// Declare the external name
			if has(defined_names, extname) {
				log.Fatalln("Error: Can not declare external symbol, name is already defined: " + extname)
			}
			// Store the name of the declared constant in defined_names
			defined_names = append(defined_names, extname)
			// Return a comment
			return "extern " + extname + "\t\t\t; external symbol\n"
		} else {
			log.Fatalln("Error: extern with invalid name:", st[1].value)
		}
	} else if (st[0].t == KEYWORD) && (st[0].value == "end") && (len(st) == 1) {
		if in_function != "" {
			// Return from the function if "end" is encountered
			ret := Token{KEYWORD, "ret"}
			newstatement := Statement{ret}
			return newstatement.String()
		} else {
			log.Fatalln("Error: Not in a function, hard to tell what should be ended with \"end\".")
		}
	} else if (st[0].t == VALID_NAME) && (len(st) == 1) {
		// Just a name, assume it's a function call
		if has(defined_names, st[0].value) {
			call := Token{KEYWORD, "call"}
			newstatement := Statement{call, st[0]}
			return newstatement.String()
		} else {
			log.Fatalln("Error: No function named:", st[0].value)
		}
		// TODO: This catches too much. Narrow and fix.
/*	} else if (st[0].t == VALID_NAME) && (st[1].t == ASSIGNMENT) {
		if !has(defined_names, st[0].value) {
			// Add the variable name to the defined names
			defined_names = append(defined_names, st[0].value)
			log.Println("Note: Declaring local variable", st[0].value, "in function", in_function)
			if len(variables[in_function]) == 0 {
				variables[in_function] = make([]string, 0, 0)
			}
			variables[in_function] = append(variables[in_function], st[0].value)
		}
		// Create a new statement, where the first token is now a VARIABLE instead of just a VALID_NAME
		newstatement := make(Statement, len(st), len(st))
		for i, t := range st {
			newstatement[i] = t
		}
		newstatement[0].t = VARIABLE
		return newstatement.String()  */
	} else if (st[0].t == VARIABLE) && (st[1].t == ASSIGNMENT) && (len(st) > 2) {
		reg := "rbp"
		if platform_bits == 32 {
			reg = "ebp"
		}
		// negative base pointer offset for local variables
		paramoffset := len(variables[in_function])-1
		offset := strconv.Itoa(paramoffset*4 + 8)
		asmcode := "\tmov [" + reg + "-" + offset + "], "
		newstatements := make(Statement, len(st)-2, len(st)-2)
		for i := 2; i < len(st); i++ {
			newstatements = append(newstatements, st[i-2])
		}
		asmcode += newstatements.String()
		asmcode += "\t\t; local variable #" + strconv.Itoa(paramoffset) + "\n"
		return asmcode
	} else if st[0].value == "const" {
		log.Fatalln("Error: Incomprehensible constant:", st.String())
	} else if st[0].t == BUILTIN {
		log.Fatalln("Error: Unhandled builtin:", st[0].value)
	} else if st[0].t == KEYWORD {
		log.Fatalln("Error: Unhandled keyword:", st[0].value)
	}
	log.Println("Error: Unfamiliar statement layout: ")
	for _, token := range []Token(st) {
		log.Print(token)
	}
	os.Exit(1)
	return ";ERROR"
}

func TokensToAssembly(tokens []Token, debug bool, debug2 bool) (string, string) {
	statement := []Token{}
	asmcode := ""
	constants := ""
	for _, token := range tokens {
		if token.t == SEP {
			if len(statement) > 0 {
				asmline := Statement(statement).String()
				if (statement[0].t == KEYWORD) && (statement[0].value == "const") {
					if strings.Contains(asmline, ":") {
						if debug {
							log.Printf("CONSTANT: \"%s\"\n", strings.Split(asmline, ":")[0])
						}
					} else {
						log.Fatalln("Error: Unfamiliar constant:", asmline)
					}
					constants += asmline + "\n"
				} else {
					asmcode += asmline + "\n"
				}
			}
			statement = []Token{}
		} else {
			statement = append(statement, token)
		}
	}
	return strings.TrimSpace(constants), asmcode
}

func add_starting_point_if_missing(asmcode string) string {
	// Check if the resulting code contains a starting point or not
	if !strings.Contains(asmcode, linker_start_function) {
		log.Printf("No %s has been defined, creating one\n", linker_start_function)
		addstring := "global " + linker_start_function + "\t\t\t; make label available to the linker\n" + linker_start_function + ":\t\t\t\t; starting point of the program\n"
		if strings.Contains(asmcode, "\nmain:") {
			log.Println("...but main has been defined, using that as starting point.")
			// Add "_start:"/"start" right after "main:"
			return strings.Replace(asmcode, "\nmain:", "\n"+addstring+"main:", 1)
		}
		return addstring + "\n" + asmcode

	}
	return asmcode
}

func add_exit_token_if_missing(tokens []Token) []Token {
	var lasttoken Token
	for i := len(tokens) - 1; i >= 0; i-- {
		if tokens[i].t == SEP {
			continue
		}
		//log.Println("LAST PROPER TOKEN", tokens[i])
		lasttoken = tokens[i]
		break
	}

	// If the last token is ret or end, all is well, return the same tokens
	if (lasttoken.t == KEYWORD) && ((lasttoken.value == "ret") || (lasttoken.value == "end")) {
		return tokens
	}

	// If the last token is exit, all is well, return the same tokens
	if (lasttoken.t == BUILTIN) && (lasttoken.value == "exit") {
		return tokens
	}

	//log.Fatalln("Error: Last token is not ret")

	// If not, add an exit statement and return

	newtokens := make([]Token, len(tokens)+2, len(tokens)+2)
	for i, _ := range tokens {
		newtokens[i] = tokens[i]
	}

	ret_token := Token{BUILTIN, "exit"}
	newtokens[len(tokens)] = ret_token

	sep_token := Token{SEP, ";"}
	newtokens[len(tokens)+1] = sep_token

	//log.Println(tokens)
	//log.Println(newtokens)

	return newtokens
}

func main() {
	name := "Battlestar"
	version := "0.1"
	log.Println(name + " compiler")
	log.Println("Version " + version)
	log.Println("Alexander Rødseth")
	log.Println("2014")
	log.Println("MIT licensed")

	// Initialize global maps and slices
	defined_names = make([]string, 0, 0)
	variables = make(map[string][]string)

	// TODO: Automatically discover 32-bit/64-bit and Linux/OS X
	// Check for -bits=32 or -bits=64 (default)
	bits := flag.Int("bits", 64, "Output 32-bit or 64-bit x86 assembly")
	// Check for -osx=true or -osx=false (default)
	is_osx := flag.Bool("osx", false, "On OS X?")

	flag.Parse()

	platform_bits = *bits
	osx = *is_osx

	// TODO: Consider adding an option for "start" as well, or a custom
	// start symbol

	if osx {
		linker_start_function = "_main"
	} else {
		linker_start_function = "_start"
	}

	// Read code from stdin and output 32-bit or 64-bit assembly code
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err == nil {
		if len(strings.TrimSpace(string(bytes))) == 0 {
			// Empty program
			log.Fatalln("Error: Empty program")
		}
		t := time.Now()
		fmt.Printf("; Generated with %s %s, at %s\n\n", name, version, t.String()[:16])
		fmt.Printf("bits %d\n", platform_bits)
		tokens := add_exit_token_if_missing(tokenize(string(bytes), true))
		log.Println("--- Done tokenizing ---")
		constants, asmcode := TokensToAssembly(tokens, true, false)
		if constants != "" {
			fmt.Println("section .data\n")
			fmt.Println(constants + "\n")
		}
		if asmcode != "" {
			fmt.Println("section .text\n")
			asmcode = add_starting_point_if_missing(asmcode)
			fmt.Println(asmcode + "\n")
		}
	}
}
