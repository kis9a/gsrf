# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This repository contains the GSRF (Go Symbol Representation Format) specification - a standardized notation for representing Go symbols (functions, methods, types) across the Go ecosystem. This is a specification document project, not an implementation.

## Repository Structure

- `spec.md` - The complete GSRF specification document defining v1.0 and v1.1 formats
- `.claude/` - Claude AI assistant configuration

## Key Concepts

### GSRF Format Versions

**v1.0** - Basic format supporting:
- Function notation: `pkg.Function`
- Method notation: `pkg.(*Type).Method` or `pkg.(Type).Method`
- Init functions: `pkg.init` (with indices for multiple)
- Anonymous functions: `pkg.Function·lit` (with indices)

**v1.1** - Extended format adding:
- Generics support: `pkg.Function[T,U]` and `pkg.(*Type[T]).Method`
- Build contexts: `pkg.Function#linux#amd64`
- Metadata: `pkg.Function@{src:file.go:12:1}`

### Related Projects

The parent directory contains related implementations:
- `calldigraph`, `calldigraph2`, `calldigraph3`, `calldigraph4` - Tools using GSRF format
- `gsrf-old`, `gsrf2` - Previous GSRF parser/lexer implementations

## Common Tasks

Since this is a specification document:
- To review the specification: Read `spec.md`
- To understand grammar: See sections 5.1 (v1.0 BNF) and 6.2 (v1.1 BNF)
- To see examples: Refer to section 9 in the spec
- To understand migration from other formats: See section 8

## Important Notes

- This repository contains only the specification, not any Go code implementation
- The specification defines how Go symbols should be represented in a standardized way
- It addresses edge cases like init functions, anonymous functions, and generics
- The format is designed to be both human-readable and machine-parsable