package parser

import (
	"fmt"

	"github.com/wudi/php-parser/ast"
	"github.com/wudi/php-parser/lexer"
)

// ============= STATEMENT PARSER INITIALIZATION =============

func (p *PrattParser) initializeStatementParsers() {
	p.statementParsers = map[lexer.TokenType]StatementParseFn{
		// Control flow statements
		lexer.T_IF:            p.parseIfStatement,
		lexer.T_WHILE:         p.parseWhileStatement,
		lexer.T_FOR:           p.parseForStatement,
		lexer.T_FOREACH:       p.parseForeachStatement,
		lexer.T_DO:            p.parseDoWhileStatement,
		lexer.T_SWITCH:        p.parseSwitchStatement,
		
		// Jump statements
		lexer.T_BREAK:         p.parseBreakStatement,
		lexer.T_CONTINUE:      p.parseContinueStatement,
		lexer.T_RETURN:        p.parseReturnStatement,
		lexer.T_GOTO:          p.parseGotoStatement,
		lexer.T_THROW:         p.parseThrowStatement,
		
		// Exception handling
		lexer.T_TRY:           p.parseTryStatement,
		
		// Variable declarations
		lexer.T_GLOBAL:        p.parseGlobalStatement,
		lexer.T_STATIC:        p.parseStaticStatement,
		lexer.T_UNSET:         p.parseUnsetStatement,
		
		// Output statements
		lexer.T_ECHO:          p.parseEchoStatement,
		lexer.T_PRINT:         p.parsePrintStatement,
		
		// Other statements
		lexer.T_DECLARE:       p.parseDeclareStatement,
		lexer.T_HALT_COMPILER: p.parseHaltCompilerStatement,
		
		// Block statement
		lexer.TOKEN_LBRACE:    p.parseBlockStatement,
		
		// Label statement (identifier followed by colon)
		lexer.T_STRING:        p.parseIdentifierOrLabelStatement,
	}
}

// ============= CONTROL FLOW STATEMENTS =============

func (p *PrattParser) parseIfStatement() ast.Statement {
	position := p.currentToken.Position
	
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}
	
	p.nextToken()
	condition := p.parseExpression(LOWEST)
	if condition == nil {
		return nil
	}
	
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}
	
	p.nextToken()
	
	// Check for alternative syntax (if: ... endif;)
	var isAlternative bool
	var thenStatement ast.Statement
	
	if p.currentTokenIs(lexer.TOKEN_COLON) {
		isAlternative = true
		p.nextToken()
		thenStatement = p.parseStatementList([]lexer.TokenType{lexer.T_ELSEIF, lexer.T_ELSE, lexer.T_ENDIF})
	} else {
		thenStatement = p.parseStatement()
	}
	
	if thenStatement == nil {
		return nil
	}
	
	var elseIfStatements []*ast.ElseIfStatement
	var elseStatement ast.Statement
	
	// Parse elseif clauses
	for p.peekTokenIs(lexer.T_ELSEIF) {
		p.nextToken() // consume ELSEIF
		
		elseIfPos := p.currentToken.Position
		
		if !p.expectPeek(lexer.TOKEN_LPAREN) {
			return nil
		}
		
		p.nextToken()
		elseIfCondition := p.parseExpression(LOWEST)
		if elseIfCondition == nil {
			return nil
		}
		
		if !p.expectPeek(lexer.TOKEN_RPAREN) {
			return nil
		}
		
		p.nextToken()
		
		var elseIfBody ast.Statement
		if isAlternative {
			if !p.expectPeek(lexer.TOKEN_COLON) {
				return nil
			}
			p.nextToken()
			elseIfBody = p.parseStatementList([]lexer.TokenType{lexer.T_ELSEIF, lexer.T_ELSE, lexer.T_ENDIF})
		} else {
			elseIfBody = p.parseStatement()
		}
		
		if elseIfBody == nil {
			return nil
		}
		
		elseIfStatements = append(elseIfStatements, &ast.ElseIfStatement{
			BaseNode: ast.BaseNode{
				Kind:     ast.ASTIfElem,
				Position: elseIfPos,
				LineNo:   elseIfPos.Line,
			},
			Condition: elseIfCondition,
			Body:      elseIfBody,
		})
	}
	
	// Parse else clause
	if p.peekTokenIs(lexer.T_ELSE) {
		p.nextToken() // consume ELSE
		p.nextToken()
		
		if isAlternative {
			if !p.expectPeek(lexer.TOKEN_COLON) {
				return nil
			}
			p.nextToken()
			elseStatement = p.parseStatementList([]lexer.TokenType{lexer.T_ENDIF})
		} else {
			elseStatement = p.parseStatement()
		}
	}
	
	// Handle alternative syntax ending
	if isAlternative {
		if !p.expectPeek(lexer.T_ENDIF) {
			return nil
		}
		if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
			return nil
		}
	}
	
	return &ast.IfStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTIf,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Condition:        condition,
		ThenStatement:    thenStatement,
		ElseIfStatements: elseIfStatements,
		ElseStatement:    elseStatement,
		IsAlternative:    isAlternative,
	}
}

func (p *PrattParser) parseWhileStatement() ast.Statement {
	position := p.currentToken.Position
	
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}
	
	p.nextToken()
	condition := p.parseExpression(LOWEST)
	if condition == nil {
		return nil
	}
	
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}
	
	p.nextToken()
	
	// Check for alternative syntax (while: ... endwhile;)
	var body ast.Statement
	var isAlternative bool
	
	if p.currentTokenIs(lexer.TOKEN_COLON) {
		isAlternative = true
		p.nextToken()
		body = p.parseStatementList([]lexer.TokenType{lexer.T_ENDWHILE})
		
		if !p.expectPeek(lexer.T_ENDWHILE) {
			return nil
		}
		if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
			return nil
		}
	} else {
		body = p.parseStatement()
	}
	
	if body == nil {
		return nil
	}
	
	return &ast.WhileStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTWhile,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Condition:     condition,
		Body:          body,
		IsAlternative: isAlternative,
	}
}

func (p *PrattParser) parseForStatement() ast.Statement {
	position := p.currentToken.Position
	
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}
	
	// Parse init expressions
	p.nextToken()
	var initExpressions []ast.Expression
	if !p.currentTokenIs(lexer.TOKEN_SEMICOLON) {
		initExpressions = p.parseExpressionList(lexer.TOKEN_SEMICOLON)
	}
	
	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}
	
	// Parse condition expressions
	p.nextToken()
	var conditionExpressions []ast.Expression
	if !p.currentTokenIs(lexer.TOKEN_SEMICOLON) {
		conditionExpressions = p.parseExpressionList(lexer.TOKEN_SEMICOLON)
	}
	
	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}
	
	// Parse update expressions
	p.nextToken()
	var updateExpressions []ast.Expression
	if !p.currentTokenIs(lexer.TOKEN_RPAREN) {
		updateExpressions = p.parseExpressionList(lexer.TOKEN_RPAREN)
	}
	
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}
	
	p.nextToken()
	
	// Check for alternative syntax (for: ... endfor;)
	var body ast.Statement
	var isAlternative bool
	
	if p.currentTokenIs(lexer.TOKEN_COLON) {
		isAlternative = true
		p.nextToken()
		body = p.parseStatementList([]lexer.TokenType{lexer.T_ENDFOR})
		
		if !p.expectPeek(lexer.T_ENDFOR) {
			return nil
		}
		if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
			return nil
		}
	} else {
		body = p.parseStatement()
	}
	
	if body == nil {
		return nil
	}
	
	return &ast.ForStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTFor,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Init:          initExpressions,
		Condition:     conditionExpressions,
		Update:        updateExpressions,
		Body:          body,
		IsAlternative: isAlternative,
	}
}

func (p *PrattParser) parseForeachStatement() ast.Statement {
	position := p.currentToken.Position
	
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}
	
	p.nextToken()
	iterable := p.parseExpression(LOWEST)
	if iterable == nil {
		return nil
	}
	
	if !p.expectPeek(lexer.T_AS) {
		return nil
	}
	
	p.nextToken()
	
	// Parse key => value or just value
	var key, value ast.Expression
	var isReference bool
	
	// Check for reference
	if p.currentTokenIs(lexer.TOKEN_AMPERSAND) {
		isReference = true
		p.nextToken()
	}
	
	firstExpr := p.parseExpression(LOWEST)
	if firstExpr == nil {
		return nil
	}
	
	// Check if we have key => value syntax
	if p.peekTokenIs(lexer.T_DOUBLE_ARROW) {
		key = firstExpr
		p.nextToken() // consume =>
		p.nextToken()
		
		// Check for reference on value
		if p.currentTokenIs(lexer.TOKEN_AMPERSAND) {
			isReference = true
			p.nextToken()
		}
		
		value = p.parseExpression(LOWEST)
		if value == nil {
			return nil
		}
	} else {
		value = firstExpr
	}
	
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}
	
	p.nextToken()
	
	// Check for alternative syntax (foreach: ... endforeach;)
	var body ast.Statement
	var isAlternative bool
	
	if p.currentTokenIs(lexer.TOKEN_COLON) {
		isAlternative = true
		p.nextToken()
		body = p.parseStatementList([]lexer.TokenType{lexer.T_ENDFOREACH})
		
		if !p.expectPeek(lexer.T_ENDFOREACH) {
			return nil
		}
		if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
			return nil
		}
	} else {
		body = p.parseStatement()
	}
	
	if body == nil {
		return nil
	}
	
	return &ast.ForeachStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTForeach,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Iterable:      iterable,
		Key:           key,
		Value:         value,
		Body:          body,
		IsReference:   isReference,
		IsAlternative: isAlternative,
	}
}

func (p *PrattParser) parseDoWhileStatement() ast.Statement {
	position := p.currentToken.Position
	
	p.nextToken()
	body := p.parseStatement()
	if body == nil {
		return nil
	}
	
	if !p.expectPeek(lexer.T_WHILE) {
		return nil
	}
	
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}
	
	p.nextToken()
	condition := p.parseExpression(LOWEST)
	if condition == nil {
		return nil
	}
	
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}
	
	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}
	
	return &ast.DoWhileStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTDoWhile,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Body:      body,
		Condition: condition,
	}
}

func (p *PrattParser) parseSwitchStatement() ast.Statement {
	position := p.currentToken.Position
	
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}
	
	p.nextToken()
	expr := p.parseExpression(LOWEST)
	if expr == nil {
		return nil
	}
	
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}
	
	p.nextToken()
	
	// Parse switch body - can be {...} or :...endswitch;
	var cases []*ast.CaseStatement
	var isAlternative bool
	
	if p.currentTokenIs(lexer.TOKEN_LBRACE) {
		p.nextToken()
		
		// Skip optional semicolon after opening brace
		if p.currentTokenIs(lexer.TOKEN_SEMICOLON) {
			p.nextToken()
		}
		
		cases = p.parseCaseList(lexer.TOKEN_RBRACE)
		
		if !p.expectPeek(lexer.TOKEN_RBRACE) {
			return nil
		}
	} else if p.currentTokenIs(lexer.TOKEN_COLON) {
		isAlternative = true
		p.nextToken()
		
		// Skip optional semicolon after colon
		if p.currentTokenIs(lexer.TOKEN_SEMICOLON) {
			p.nextToken()
		}
		
		cases = p.parseCaseList(lexer.T_ENDSWITCH)
		
		if !p.expectPeek(lexer.T_ENDSWITCH) {
			return nil
		}
		if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
			return nil
		}
	} else {
		p.errors = append(p.errors, "expected '{' or ':' after switch expression")
		return nil
	}
	
	return &ast.SwitchStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTSwitch,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression:    expr,
		Cases:         cases,
		IsAlternative: isAlternative,
	}
}

// ============= JUMP STATEMENTS =============

func (p *PrattParser) parseBreakStatement() ast.Statement {
	position := p.currentToken.Position
	
	p.nextToken()
	
	var level ast.Expression
	if !p.currentTokenIs(lexer.TOKEN_SEMICOLON) {
		level = p.parseExpression(LOWEST)
	}
	
	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}
	
	return &ast.BreakStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTBreak,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Level: level,
	}
}

func (p *PrattParser) parseContinueStatement() ast.Statement {
	position := p.currentToken.Position
	
	p.nextToken()
	
	var level ast.Expression
	if !p.currentTokenIs(lexer.TOKEN_SEMICOLON) {
		level = p.parseExpression(LOWEST)
	}
	
	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}
	
	return &ast.ContinueStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTContinue,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Level: level,
	}
}

func (p *PrattParser) parseReturnStatement() ast.Statement {
	position := p.currentToken.Position
	
	p.nextToken()
	
	var value ast.Expression
	if !p.currentTokenIs(lexer.TOKEN_SEMICOLON) {
		value = p.parseExpression(LOWEST)
	}
	
	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}
	
	return &ast.ReturnStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTReturn,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Value: value,
	}
}

func (p *PrattParser) parseGotoStatement() ast.Statement {
	position := p.currentToken.Position
	
	if !p.expectPeek(lexer.T_STRING) {
		return nil
	}
	
	label := p.currentToken.Value
	
	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}
	
	return &ast.GotoStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTGoto,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Label: label,
	}
}

func (p *PrattParser) parseThrowStatement() ast.Statement {
	position := p.currentToken.Position
	
	p.nextToken()
	expression := p.parseExpression(LOWEST)
	if expression == nil {
		return nil
	}
	
	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}
	
	return &ast.ThrowStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTThrow,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression: expression,
	}
}

// ============= EXCEPTION HANDLING =============

func (p *PrattParser) parseTryStatement() ast.Statement {
	position := p.currentToken.Position
	
	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}
	
	p.nextToken()
	tryBlock := p.parseStatementList([]lexer.TokenType{lexer.TOKEN_RBRACE})
	
	if !p.expectPeek(lexer.TOKEN_RBRACE) {
		return nil
	}
	
	// Parse catch clauses
	var catchClauses []*ast.CatchClause
	for p.peekTokenIs(lexer.T_CATCH) {
		p.nextToken() // consume CATCH
		
		catchPos := p.currentToken.Position
		
		if !p.expectPeek(lexer.TOKEN_LPAREN) {
			return nil
		}
		
		p.nextToken()
		
		// Parse exception types (can be multiple with | separator)
		var exceptionTypes []ast.Expression
		exceptionTypes = append(exceptionTypes, p.parseExpression(LOWEST))
		
		for p.peekTokenIs(lexer.TOKEN_PIPE) {
			p.nextToken() // consume |
			p.nextToken()
			exceptionTypes = append(exceptionTypes, p.parseExpression(LOWEST))
		}
		
		// Parse optional variable
		var variable ast.Expression
		if p.peekTokenIs(lexer.T_VARIABLE) {
			p.nextToken()
			variable = p.parseVariable()
		}
		
		if !p.expectPeek(lexer.TOKEN_RPAREN) {
			return nil
		}
		
		if !p.expectPeek(lexer.TOKEN_LBRACE) {
			return nil
		}
		
		p.nextToken()
		catchBody := p.parseStatementList([]lexer.TokenType{lexer.TOKEN_RBRACE})
		
		if !p.expectPeek(lexer.TOKEN_RBRACE) {
			return nil
		}
		
		catchClauses = append(catchClauses, &ast.CatchClause{
			BaseNode: ast.BaseNode{
				Kind:     ast.ASTCatch,
				Position: catchPos,
				LineNo:   catchPos.Line,
			},
			ExceptionTypes: exceptionTypes,
			Variable:       variable,
			Body:           catchBody,
		})
	}
	
	// Parse optional finally clause
	var finallyClause *ast.FinallyClause
	if p.peekTokenIs(lexer.T_FINALLY) {
		p.nextToken() // consume FINALLY
		
		finallyPos := p.currentToken.Position
		
		if !p.expectPeek(lexer.TOKEN_LBRACE) {
			return nil
		}
		
		p.nextToken()
		finallyBody := p.parseStatementList([]lexer.TokenType{lexer.TOKEN_RBRACE})
		
		if !p.expectPeek(lexer.TOKEN_RBRACE) {
			return nil
		}
		
		finallyClause = &ast.FinallyClause{
			BaseNode: ast.BaseNode{
				Kind:     ast.ASTFinally,
				Position: finallyPos,
				LineNo:   finallyPos.Line,
			},
			Body: finallyBody,
		}
	}
	
	if len(catchClauses) == 0 && finallyClause == nil {
		p.errors = append(p.errors, "try statement must have at least one catch or finally clause")
		return nil
	}
	
	return &ast.TryStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTTry,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		TryBlock:      tryBlock,
		CatchClauses:  catchClauses,
		FinallyClause: finallyClause,
	}
}

// ============= UTILITY FUNCTIONS =============

// parseStatementList parses a list of statements until one of the stop tokens is encountered
func (p *PrattParser) parseStatementList(stopTokens []lexer.TokenType) ast.Statement {
	var statements []ast.Statement
	
	for !p.currentTokenIs(lexer.T_EOF) && !p.tokenInList(p.currentToken.Type, stopTokens) {
		stmt := p.parseStatement()
		if stmt != nil {
			statements = append(statements, stmt)
		}
		p.nextToken()
	}
	
	if len(statements) == 1 {
		return statements[0]
	}
	
	return &ast.BlockStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTStmtList,
			Position: statements[0].GetPosition(),
			LineNo:   statements[0].GetLineNo(),
		},
		Statements: statements,
	}
}

// parseExpressionList parses a comma-separated list of expressions
func (p *PrattParser) parseExpressionList(stopToken lexer.TokenType) []ast.Expression {
	var expressions []ast.Expression
	
	if p.currentTokenIs(stopToken) {
		return expressions
	}
	
	expressions = append(expressions, p.parseExpression(LOWEST))
	
	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken() // consume comma
		p.nextToken()
		expressions = append(expressions, p.parseExpression(LOWEST))
	}
	
	return expressions
}

// parseCaseList parses a list of case statements
func (p *PrattParser) parseCaseList(stopToken lexer.TokenType) []*ast.CaseStatement {
	var cases []*ast.CaseStatement
	
	for !p.currentTokenIs(lexer.T_EOF) && !p.currentTokenIs(stopToken) {
		if p.currentTokenIs(lexer.T_CASE) {
			casePos := p.currentToken.Position
			
			p.nextToken()
			condition := p.parseExpression(LOWEST)
			if condition == nil {
				break
			}
			
			// Expect : or ;
			if p.peekTokenIs(lexer.TOKEN_COLON) {
				p.nextToken()
			} else if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
				p.nextToken()
			} else {
				p.errors = append(p.errors, "expected ':' or ';' after case condition")
				break
			}
			
			p.nextToken()
			
			// Parse statements until next case/default/end
			var statements []ast.Statement
			for !p.currentTokenIs(lexer.T_CASE) && !p.currentTokenIs(lexer.T_DEFAULT) && 
				!p.currentTokenIs(stopToken) && !p.currentTokenIs(lexer.T_EOF) {
				
				stmt := p.parseStatement()
				if stmt != nil {
					statements = append(statements, stmt)
				}
				p.nextToken()
			}
			
			cases = append(cases, &ast.CaseStatement{
				BaseNode: ast.BaseNode{
					Kind:     ast.ASTSwitchCase,
					Position: casePos,
					LineNo:   casePos.Line,
				},
				Condition:  condition,
				Statements: statements,
			})
			
		} else if p.currentTokenIs(lexer.T_DEFAULT) {
			defaultPos := p.currentToken.Position
			
			// Expect : or ;
			if p.peekTokenIs(lexer.TOKEN_COLON) {
				p.nextToken()
			} else if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
				p.nextToken()
			} else {
				p.errors = append(p.errors, "expected ':' or ';' after default")
				break
			}
			
			p.nextToken()
			
			// Parse statements until next case/default/end
			var statements []ast.Statement
			for !p.currentTokenIs(lexer.T_CASE) && !p.currentTokenIs(lexer.T_DEFAULT) && 
				!p.currentTokenIs(stopToken) && !p.currentTokenIs(lexer.T_EOF) {
				
				stmt := p.parseStatement()
				if stmt != nil {
					statements = append(statements, stmt)
				}
				p.nextToken()
			}
			
			cases = append(cases, &ast.CaseStatement{
				BaseNode: ast.BaseNode{
					Kind:     ast.ASTSwitchCase,
					Position: defaultPos,
					LineNo:   defaultPos.Line,
				},
				Condition:  nil, // nil indicates default case
				Statements: statements,
			})
			
		} else {
			p.nextToken()
		}
	}
	
	return cases
}

// tokenInList checks if a token type is in the given list
func (p *PrattParser) tokenInList(tokenType lexer.TokenType, tokens []lexer.TokenType) bool {
	for _, t := range tokens {
		if tokenType == t {
			return true
		}
	}
	return false
}

// ============= BLOCK AND EXPRESSION STATEMENTS =============

func (p *PrattParser) parseBlockStatement() ast.Statement {
	position := p.currentToken.Position
	
	p.nextToken()
	statements := p.parseStatementList([]lexer.TokenType{lexer.TOKEN_RBRACE})
	
	if !p.expectPeek(lexer.TOKEN_RBRACE) {
		return nil
	}
	
	if block, ok := statements.(*ast.BlockStatement); ok {
		return block
	}
	
	return &ast.BlockStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTStmtList,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Statements: []ast.Statement{statements},
	}
}

func (p *PrattParser) parseExpressionStatement() ast.Statement {
	position := p.currentToken.Position
	
	expression := p.parseExpression(LOWEST)
	if expression == nil {
		return nil
	}
	
	// Optional semicolon for expression statements
	if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
		p.nextToken()
	}
	
	return &ast.ExpressionStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTExprStmt,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression: expression,
	}
}

// parseIdentifierOrLabelStatement handles both identifiers in expressions and labels
func (p *PrattParser) parseIdentifierOrLabelStatement() ast.Statement {
	// Check if this is a label (identifier followed by colon)
	if p.peekTokenIs(lexer.TOKEN_COLON) {
		position := p.currentToken.Position
		label := p.currentToken.Value
		
		p.nextToken() // consume identifier
		p.nextToken() // consume colon
		
		return &ast.LabelStatement{
			BaseNode: ast.BaseNode{
				Kind:     ast.ASTLabel,
				Position: position,
				LineNo:   uint32(position.Line),
			},
			Name: label,
		}
	}
	
	// Otherwise, parse as expression statement
	return p.parseExpressionStatement()
}

// ============= MISSING STATEMENT PARSING METHODS =============

func (p *PrattParser) parseGlobalStatement() ast.Statement {
	// TODO: Implement global statement parsing
	position := p.currentToken.Position
	p.nextToken() // Skip for now
	return &ast.ExpressionStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTExprStmt,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression: nil, // Placeholder
	}
}

func (p *PrattParser) parseStaticStatement() ast.Statement {
	// TODO: Implement static statement parsing
	position := p.currentToken.Position
	p.nextToken() // Skip for now
	return &ast.ExpressionStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTExprStmt,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression: nil, // Placeholder
	}
}

func (p *PrattParser) parseUnsetStatement() ast.Statement {
	// TODO: Implement unset statement parsing
	position := p.currentToken.Position
	p.nextToken() // Skip for now
	return &ast.ExpressionStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTExprStmt,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression: nil, // Placeholder
	}
}

func (p *PrattParser) parseEchoStatement() ast.Statement {
	// TODO: Implement echo statement parsing
	position := p.currentToken.Position
	p.nextToken() // Skip for now
	return &ast.ExpressionStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTExprStmt,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression: nil, // Placeholder
	}
}

func (p *PrattParser) parsePrintStatement() ast.Statement {
	// TODO: Implement print statement parsing
	position := p.currentToken.Position
	p.nextToken() // Skip for now
	return &ast.ExpressionStatement{
		BaseNode: ast.BaseNode{
			Kind:     ast.ASTExprStmt,
			Position: position,
			LineNo:   uint32(position.Line),
		},
		Expression: nil, // Placeholder
	}
}