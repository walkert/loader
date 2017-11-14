package main

import (
	"testing"
)

func TestOpen(t *testing.T)            { testOpen(t) }
func TestOpenWithVars(t *testing.T)    { testOpenWithVars(t) }
func TestOpenWithVarsRW(t *testing.T)  { testOpenWithVarsRW(t) }
func TestFOpen(t *testing.T)           { testFOpen(t) }
func TestFOpenWithVars(t *testing.T)   { testFOpenWithVars(t) }
func TestFOpenWithVarsRW(t *testing.T) { testFOpenWithVarsRW(t) }
func TestDebugFailure(t *testing.T)    { testDebugFailure(t) }
