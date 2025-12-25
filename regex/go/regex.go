package main

import (
	"fmt"
	"strconv"
	"strings"
)

type tokenType uint8

const (
	group           tokenType = iota
	bracket         tokenType = iota
	or              tokenType = iota
	repeat          tokenType = iota
	literal         tokenType = iota
	groupUncaptured tokenType = iota
)

type token struct {
	tokenType tokenType
	value     interface{}
}

type parseContext struct {
	pos    int
	tokens []token
}

type repeatPayload struct {
	min   int
	max   int
	token token
}

func parse(regex string) *parseContext {
	ctx := &parseContext{
		pos:    0,
		tokens: []token{},
	}
	for ctx.pos < len(regex) {
		process(regex, ctx)
		ctx.pos++
	}
	return ctx
}

func process(regex string, ctx *parseContext) {
	ch := regex[ctx.pos]
	switch ch {
	case '(':
		groupCtx := &parseContext{
			pos:    ctx.pos,
			tokens: []token{},
		}
		parseGroup(regex, groupCtx)
		ctx.tokens = append(ctx.tokens, token{
			tokenType: group,
			value:     groupCtx.tokens,
		})
	case '[':
		parseBracket(regex, ctx)
	case '|':
		parseOr(regex, ctx)
	case '*' | '?' | '+':
		parseRepeat(regex, ctx)
	case '{':
		parseRepeatSpecified(regex, ctx)
	default:
		t := token{
			tokenType: literal,
			value:     ch,
		}
		ctx.tokens = append(ctx.tokens, t)
	}
}

func parseGroup(regex string, ctx *parseContext) {
	ctx.pos += 1
	for regex[ctx.pos] != ')' {
		process(regex, ctx)
		ctx.pos += 1
	}
}

func parseBracket(regex string, ctx *parseContext) {
	ctx.pos++
	var literals []string
	for regex[ctx.pos] != ']' {
		ch := regex[ctx.pos]
		if ch == '-' {
			next := regex[ctx.pos+1]
			previous := literals[len(literals)-1][0]
			literals[len(literals)-1] = fmt.Sprintf("%c%c", previous, next)
			ctx.pos++
		} else {
			literals = append(literals, fmt.Sprintf("%c", ch))
		}
		ctx.pos++
	}
	literalsSet := map[uint8]bool{}
	for _, l := range literals {
		for i := l[0]; i <= l[len(l)-1]; i++ {
			literalsSet[i] = true
		}
	}
	ctx.tokens = append(ctx.tokens, token{
		tokenType: bracket,
		value:     literalsSet,
	})
}

func parseOr(regex string, ctx *parseContext) {
	rhsContext := &parseContext{
		pos:    ctx.pos,
		tokens: []token{},
	}
	rhsContext.pos += 1
	for rhsContext.pos < len(regex) && regex[rhsContext.pos] != ')' {
		process(regex, rhsContext)
		rhsContext.pos++
	}
	left := token{
		tokenType: groupUncaptured,
		value:     ctx.tokens,
	}
	right := token{
		tokenType: groupUncaptured,
		value:     rhsContext.tokens,
	}
	ctx.pos = rhsContext.pos

	ctx.tokens = []token{{
		tokenType: or,
		value:     []token{left, right},
	}}

}

const repeatInfinity = -1

func parseRepeat(regex string, ctx *parseContext) {
	ch := regex[ctx.pos]
	var min, max int
	if ch == '*' {
		min = 0
		max = repeatInfinity
	}
	if ch == '?' {
		min = 0
		max = 1
	}
	if ch == '+' {
		min = 1
		max = repeatInfinity
	}
	lastToken := ctx.tokens[len(ctx.tokens)-1]
	ctx.tokens[len(ctx.tokens)-1] = token{
		tokenType: repeat,
		value: repeatPayload{
			min:   min,
			max:   max,
			token: lastToken,
		},
	}

}

func parseRepeatSpecified(regex string, ctx *parseContext) {
	start := ctx.pos + 1
	for regex[ctx.pos] != '}' {
		ctx.pos++
	}
	boundariesStr := regex[start:ctx.pos]
	pieces := strings.Split(boundariesStr, ",")
	var min, max int
	if len(pieces) == 1 {
		if value, err := strconv.Atoi(pieces[0]); err != nil {
			panic(err.Error())
		} else {
			min = value
			max = value
		}
	} else if len(pieces) == 2 {
		if value, err := strconv.Atoi(pieces[0]); err != nil {
			panic(err.Error())
		} else {
			min = value
		}
		if pieces[1] == "" {
			max = repeatInfinity
		} else if value, err := strconv.Atoi(pieces[1]); err != nil {
			panic(err.Error())
		} else {
			max = value
		}
	} else {
		panic(fmt.Sprintf("There must be either 1 or 2 values specified for this quantifier. provided is '%s'", boundariesStr))
	}
	lastToken := ctx.tokens[len(ctx.tokens)-1]
	ctx.tokens[len(ctx.tokens)-1] = token{
		tokenType: repeat,
		value: repeatPayload{
			min:   min,
			max:   max,
			token: lastToken,
		},
	}
}

type state struct {
	start       bool
	terminal    bool
	transitions map[uint8][]*state
}

func tokenToNfa(t *token) (*state, *state) {
	start := &state{
		transitions: map[uint8][]*state{},
	}
	end := &state{
		transitions: map[uint8][]*state{},
	}
	switch t.tokenType {
	case literal:
		ch := t.value.(uint8)
		start.transitions[ch] = []*state{end}
	case or:
		values := t.value.([]token)
		left := values[0]
		right := values[1]
		s1, e1 := tokenToNfa(&left)
		s2, e2 := tokenToNfa(&right)
		start.transitions[epsilonChar] = []*state{s1, s2}
		e1.transitions[epsilonChar] = []*state{end}
		e2.transitions[epsilonChar] = []*state{end}
	case bracket:
		literals := t.value.(map[uint8]bool)
		for l := range literals {
			start.transitions[l] = []*state{end}
		}
	case group, groupUncaptured:
		tokens := t.value.([]token)
		start, end = tokenToNfa(&tokens[0])
		for i := 1; i < len(tokens); i++ {
			ts, te := tokenToNfa(&tokens[i])
			end.transitions[epsilonChar] = append(end.transitions[epsilonChar], ts)
			end = te
		}
	case repeat:
	default:
		panic("unkown type of token")
	}
	return start, end
}

const epsilonChar uint8 = 0

func toNfa(ctx *parseContext) *state {
	startState, endState := tokenToNfa(&ctx.tokens[0])
	for i := 0; i < len(ctx.tokens); i++ {
		startNext, endNext := tokenToNfa(&ctx.tokens[i])
		endState.transitions[epsilonChar] = append(endState.transitions[epsilonChar], startNext)
		endState = endNext
	}
	start := &state{
		transitions: map[uint8][]*state{
			epsilonChar: {startState},
		},
		start: true,
	}
	end := &state{
		transitions: map[uint8][]*state{},
		terminal:    true,
	}
	endState.transitions[epsilonChar] = append(endState.transitions[epsilonChar], end)

	return start
}
