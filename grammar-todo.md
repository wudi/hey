# PHP Grammar Analysis & Implementation Todo List

Based on analysis of `/home/ubuntu/php-src/Zend/zend_language_parser.y` (lines 298-1674)

## Core Grammar Rules Extracted

### 1. Program Structure
- **start** → top_statement_list
- **top_statement_list** → top_statement_list top_statement | ε
- **inner_statement_list** → inner_statement_list inner_statement | ε
- **top_statement** → statement | attributed_top_statement | attributes attributed_top_statement | T_HALT_COMPILER | namespace declarations | use declarations
- **inner_statement** → statement | attributed_statement | attributes attributed_statement | T_HALT_COMPILER (error)

### 2. Identifiers & Names
- **identifier** → T_STRING | semi_reserved
- **reserved_non_modifiers** → T_INCLUDE | T_INCLUDE_ONCE | T_EVAL | ... (50+ keywords)
- **semi_reserved** → reserved_non_modifiers | T_STATIC | T_ABSTRACT | T_FINAL | visibility modifiers
- **namespace_declaration_name** → identifier | T_NAME_QUALIFIED
- **namespace_name** → T_STRING | T_NAME_QUALIFIED
- **legacy_namespace_name** → namespace_name | T_NAME_FULLY_QUALIFIED
- **name** → T_STRING | T_NAME_QUALIFIED | T_NAME_FULLY_QUALIFIED | T_NAME_RELATIVE

### 3. Attributes System (PHP 8.0+)
- **attribute_decl** → class_name | class_name argument_list
- **attribute_group** → attribute_decl | attribute_group ',' attribute_decl
- **attribute** → T_ATTRIBUTE attribute_group possible_comma ']'
- **attributes** → attribute | attributes attribute

### 4. Statements
- **statement** → compound_statement | if_stmt | alt_if_stmt | while_stmt | do_while_stmt | for_stmt | switch_stmt | break_stmt | continue_stmt | return_stmt | global_stmt | static_stmt | echo_stmt | expr_stmt | unset_stmt | foreach_stmt | declare_stmt | try_stmt | goto_stmt | label_stmt | void_cast_stmt
- **compound_statement** → '{' inner_statement_list '}'
- **expr_stmt** → expr ';'

### 5. Control Structures

#### If Statements
- **if_stmt** → if_stmt_without_else | if_stmt_without_else T_ELSE statement
- **if_stmt_without_else** → T_IF '(' expr ')' statement | if_stmt_without_else T_ELSEIF '(' expr ')' statement
- **alt_if_stmt** → alt_if_stmt_without_else T_ENDIF ';' | alt_if_stmt_without_else T_ELSE ':' inner_statement_list T_ENDIF ';'
- **alt_if_stmt_without_else** → T_IF '(' expr ')' ':' inner_statement_list | alt_if_stmt_without_else T_ELSEIF '(' expr ')' ':' inner_statement_list

#### Loop Statements
- **while_stmt** → T_WHILE '(' expr ')' while_statement
- **while_statement** → statement | ':' inner_statement_list T_ENDWHILE ';'
- **do_while_stmt** → T_DO statement T_WHILE '(' expr ')' ';'
- **for_stmt** → T_FOR '(' for_exprs ';' for_cond_exprs ';' for_exprs ')' for_statement
- **for_statement** → statement | ':' inner_statement_list T_ENDFOR ';'
- **foreach_stmt** → T_FOREACH '(' expr T_AS foreach_variable ')' foreach_statement | T_FOREACH '(' expr T_AS foreach_variable T_DOUBLE_ARROW foreach_variable ')' foreach_statement
- **foreach_statement** → statement | ':' inner_statement_list T_ENDFOREACH ';'
- **foreach_variable** → variable | ampersand variable | T_LIST '(' array_pair_list ')' | '[' array_pair_list ']'

#### Switch Statement
- **switch_stmt** → T_SWITCH '(' expr ')' switch_case_list
- **switch_case_list** → '{' case_list '}' | '{' ';' case_list '}' | ':' case_list T_ENDSWITCH ';' | ':' ';' case_list T_ENDSWITCH ';'
- **case_list** → ε | case_list T_CASE expr ':' inner_statement_list | case_list T_CASE expr ';' inner_statement_list | case_list T_DEFAULT ':' inner_statement_list | case_list T_DEFAULT ';' inner_statement_list

#### Match Expression (PHP 8.0+)
- **match** → T_MATCH '(' expr ')' '{' match_arm_list '}'
- **match_arm_list** → ε | non_empty_match_arm_list possible_comma
- **non_empty_match_arm_list** → match_arm | non_empty_match_arm_list ',' match_arm
- **match_arm** → match_arm_cond_list possible_comma T_DOUBLE_ARROW expr | T_DEFAULT possible_comma T_DOUBLE_ARROW expr
- **match_arm_cond_list** → expr | match_arm_cond_list ',' expr

#### Exception Handling
- **try_stmt** → T_TRY '{' inner_statement_list '}' catch_list finally_statement
- **catch_list** → ε | catch_list T_CATCH '(' catch_name_list optional_variable ')' '{' inner_statement_list '}'
- **catch_name_list** → class_name | catch_name_list '|' class_name
- **optional_variable** → ε | T_VARIABLE
- **finally_statement** → ε | T_FINALLY '{' inner_statement_list '}'

### 6. Class Declarations

#### Class Structure
- **class_declaration_statement** → class_modifiers T_CLASS T_STRING extends_from implements_list backup_doc_comment '{' class_statement_list '}' | T_CLASS T_STRING extends_from implements_list backup_doc_comment '{' class_statement_list '}'
- **class_modifiers** → class_modifier | class_modifiers class_modifier
- **class_modifier** → T_ABSTRACT | T_FINAL | T_READONLY
- **extends_from** → ε | T_EXTENDS class_name
- **implements_list** → ε | T_IMPLEMENTS class_name_list
- **class_name_list** → class_name | class_name_list ',' class_name

#### Trait & Interface
- **trait_declaration_statement** → T_TRAIT T_STRING backup_doc_comment '{' class_statement_list '}'
- **interface_declaration_statement** → T_INTERFACE T_STRING interface_extends_list backup_doc_comment '{' class_statement_list '}'
- **interface_extends_list** → ε | T_EXTENDS class_name_list

#### Enum (PHP 8.1+)
- **enum_declaration_statement** → T_ENUM T_STRING enum_backing_type implements_list backup_doc_comment '{' class_statement_list '}'
- **enum_backing_type** → ε | ':' type_expr
- **enum_case** → T_CASE backup_doc_comment identifier enum_case_expr ';'
- **enum_case_expr** → ε | '=' expr

#### Class Members
- **class_statement_list** → class_statement_list class_statement | ε
- **class_statement** → attributed_class_statement | attributes attributed_class_statement | T_USE class_name_list trait_adaptations
- **attributed_class_statement** → property_declaration | const_declaration | method_declaration | enum_case

##### Properties
- **property_declaration** → property_modifiers optional_type_without_static property_list ';' | property_modifiers optional_type_without_static hooked_property
- **property_modifiers** → non_empty_member_modifiers | T_VAR
- **property_list** → property_list ',' property | property
- **property** → T_VARIABLE backup_doc_comment | T_VARIABLE '=' expr backup_doc_comment
- **hooked_property** → T_VARIABLE backup_doc_comment '{' property_hook_list '}' | T_VARIABLE '=' expr backup_doc_comment '{' property_hook_list '}'

##### Property Hooks (PHP 8.4+)
- **property_hook_list** → ε | property_hook_list property_hook | property_hook_list attributes property_hook
- **optional_property_hook_list** → ε | '{' property_hook_list '}'
- **property_hook** → property_hook_modifiers returns_ref T_STRING backup_doc_comment optional_parameter_list backup_fn_flags property_hook_body backup_fn_flags
- **property_hook_modifiers** → ε | non_empty_member_modifiers
- **property_hook_body** → ';' | '{' inner_statement_list '}' | T_DOUBLE_ARROW expr ';'
- **optional_parameter_list** → ε | '(' parameter_list ')'

##### Constants
- **const_declaration** → class_const_modifiers T_CONST class_const_list ';' | class_const_modifiers T_CONST type_expr class_const_list ';'
- **class_const_modifiers** → ε | non_empty_member_modifiers
- **class_const_list** → class_const_list ',' class_const_decl | class_const_decl
- **class_const_decl** → T_STRING '=' expr backup_doc_comment | semi_reserved '=' expr backup_doc_comment

##### Methods
- **method_declaration** → method_modifiers function returns_ref identifier backup_doc_comment '(' parameter_list ')' return_type backup_fn_flags method_body backup_fn_flags
- **method_modifiers** → ε | non_empty_member_modifiers
- **method_body** → ';' | '{' inner_statement_list '}'

##### Member Modifiers
- **non_empty_member_modifiers** → member_modifier | non_empty_member_modifiers member_modifier
- **member_modifier** → T_PUBLIC | T_PROTECTED | T_PRIVATE | T_PUBLIC_SET | T_PROTECTED_SET | T_PRIVATE_SET | T_STATIC | T_ABSTRACT | T_FINAL | T_READONLY

#### Trait Adaptations
- **trait_adaptations** → ';' | '{' '}' | '{' trait_adaptation_list '}'
- **trait_adaptation_list** → trait_adaptation | trait_adaptation_list trait_adaptation
- **trait_adaptation** → trait_precedence ';' | trait_alias ';'
- **trait_precedence** → absolute_trait_method_reference T_INSTEADOF class_name_list
- **trait_alias** → trait_method_reference T_AS T_STRING | trait_method_reference T_AS reserved_non_modifiers | trait_method_reference T_AS member_modifier identifier | trait_method_reference T_AS member_modifier
- **trait_method_reference** → identifier | absolute_trait_method_reference
- **absolute_trait_method_reference** → class_name T_PAAMAYIM_NEKUDOTAYIM identifier

### 7. Function Declarations
- **function_declaration_statement** → function returns_ref function_name backup_doc_comment '(' parameter_list ')' return_type backup_fn_flags '{' inner_statement_list '}' backup_fn_flags
- **function_name** → T_STRING | T_READONLY
- **returns_ref** → ε | ampersand
- **ampersand** → T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG | T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG

#### Parameters
- **parameter_list** → non_empty_parameter_list possible_comma | ε
- **non_empty_parameter_list** → attributed_parameter | non_empty_parameter_list ',' attributed_parameter
- **attributed_parameter** → attributes parameter | parameter
- **parameter** → optional_cpp_modifiers optional_type_without_static is_reference is_variadic T_VARIABLE backup_doc_comment optional_property_hook_list | optional_cpp_modifiers optional_type_without_static is_reference is_variadic T_VARIABLE backup_doc_comment '=' expr optional_property_hook_list
- **optional_cpp_modifiers** → ε | non_empty_member_modifiers
- **is_reference** → ε | T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG
- **is_variadic** → ε | T_ELLIPSIS

#### Types
- **return_type** → ε | ':' type_expr
- **type_expr** → type | '?' type | union_type | intersection_type
- **type** → type_without_static | T_STATIC
- **type_without_static** → T_ARRAY | T_CALLABLE | name
- **optional_type_without_static** → ε | type_expr_without_static
- **type_expr_without_static** → type_without_static | '?' type_without_static | union_type_without_static | intersection_type_without_static

##### Union Types (PHP 8.0+)
- **union_type** → union_type_element '|' union_type_element | union_type '|' union_type_element
- **union_type_element** → type | '(' intersection_type ')'
- **union_type_without_static** → union_type_without_static_element '|' union_type_without_static_element | union_type_without_static '|' union_type_without_static_element
- **union_type_without_static_element** → type_without_static | '(' intersection_type_without_static ')'

##### Intersection Types (PHP 8.1+)
- **intersection_type** → type T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG type | intersection_type T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG type
- **intersection_type_without_static** → type_without_static T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG type_without_static | intersection_type_without_static T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG type_without_static

### 8. Anonymous Classes & Functions
- **anonymous_class** → anonymous_class_modifiers_optional T_CLASS ctor_arguments extends_from implements_list backup_doc_comment '{' class_statement_list '}'
- **anonymous_class_modifiers_optional** → ε | anonymous_class_modifiers
- **anonymous_class_modifiers** → class_modifier | anonymous_class_modifiers class_modifier
- **inline_function** → function returns_ref backup_doc_comment '(' parameter_list ')' lexical_vars return_type backup_fn_flags '{' inner_statement_list '}' backup_fn_flags | fn returns_ref backup_doc_comment '(' parameter_list ')' return_type T_DOUBLE_ARROW backup_fn_flags backup_lex_pos expr backup_fn_flags
- **lexical_vars** → ε | T_USE '(' lexical_var_list possible_comma ')'
- **lexical_var_list** → lexical_var_list ',' lexical_var | lexical_var
- **lexical_var** → T_VARIABLE | ampersand T_VARIABLE

### 9. Expressions

#### Primary Expressions
- **expr** → variable | assignment_expr | clone_expr | arithmetic_expr | logical_expr | comparison_expr | bitwise_expr | conditional_expr | cast_expr | new_expr | function_call | array_expr | scalar | match | inline_function | attributes inline_function | T_STATIC inline_function | attributes T_STATIC inline_function

#### Assignment Expressions
- **assignment_expr** → T_LIST '(' array_pair_list ')' '=' expr | '[' array_pair_list ']' '=' expr | variable '=' expr | variable '=' ampersand variable | variable compound_assignment_op expr
- **compound_assignment_op** → T_PLUS_EQUAL | T_MINUS_EQUAL | T_MUL_EQUAL | T_POW_EQUAL | T_DIV_EQUAL | T_CONCAT_EQUAL | T_MOD_EQUAL | T_AND_EQUAL | T_OR_EQUAL | T_XOR_EQUAL | T_SL_EQUAL | T_SR_EQUAL | T_COALESCE_EQUAL

#### Increment/Decrement
- **inc_dec_expr** → variable T_INC | T_INC variable | variable T_DEC | T_DEC variable

#### Binary Expressions
- **logical_expr** → expr T_BOOLEAN_OR expr | expr T_BOOLEAN_AND expr | expr T_LOGICAL_OR expr | expr T_LOGICAL_AND expr | expr T_LOGICAL_XOR expr
- **bitwise_expr** → expr '|' expr | expr T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG expr | expr T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG expr | expr '^' expr
- **arithmetic_expr** → expr '+' expr | expr '-' expr | expr '*' expr | expr T_POW expr | expr '/' expr | expr '%' expr | expr T_SL expr | expr T_SR expr
- **comparison_expr** → expr T_IS_IDENTICAL expr | expr T_IS_NOT_IDENTICAL expr | expr T_IS_EQUAL expr | expr T_IS_NOT_EQUAL expr | expr '<' expr | expr T_IS_SMALLER_OR_EQUAL expr | expr '>' expr | expr T_IS_GREATER_OR_EQUAL expr | expr T_SPACESHIP expr
- **instanceof_expr** → expr T_INSTANCEOF class_name_reference

#### Unary Expressions
- **unary_expr** → '+' expr | '-' expr | '!' expr | '~' expr | '@' expr

#### Conditional & Coalescing
- **conditional_expr** → expr '?' expr ':' expr | expr '?' ':' expr
- **coalescing_expr** → expr T_COALESCE expr

#### Pipe Expression (PHP 8.4+)
- **pipe_expr** → expr T_PIPE expr

#### Cast Expressions
- **cast_expr** → T_INT_CAST expr | T_DOUBLE_CAST expr | T_STRING_CAST expr | T_ARRAY_CAST expr | T_OBJECT_CAST expr | T_BOOL_CAST expr | T_UNSET_CAST expr

#### New Expression
- **new_dereferenceable** → T_NEW class_name_reference argument_list | T_NEW anonymous_class | T_NEW attributes anonymous_class
- **new_non_dereferenceable** → T_NEW class_name_reference

#### Clone Expression
- **clone_expr** → T_CLONE clone_argument_list | T_CLONE expr

#### Special Expressions
- **yield_expr** → T_YIELD | T_YIELD expr | T_YIELD expr T_DOUBLE_ARROW expr | T_YIELD_FROM expr
- **throw_expr** → T_THROW expr
- **print_expr** → T_PRINT expr
- **exit_expr** → T_EXIT ctor_arguments
- **shell_exec_expr** → '`' backticks_expr '`'
- **silence_expr** → '@' expr

### 10. Variables & References
- **variable** → callable_variable | static_member | array_object_dereferenceable T_OBJECT_OPERATOR property_name | array_object_dereferenceable T_NULLSAFE_OBJECT_OPERATOR property_name
- **simple_variable** → T_VARIABLE | '$' '{' expr '}' | '$' simple_variable
- **callable_variable** → simple_variable | array_object_dereferenceable '[' optional_expr ']' | array_object_dereferenceable T_OBJECT_OPERATOR property_name argument_list | array_object_dereferenceable T_NULLSAFE_OBJECT_OPERATOR property_name argument_list | function_call
- **new_variable** → simple_variable | new_variable '[' optional_expr ']' | new_variable T_OBJECT_OPERATOR property_name | new_variable T_NULLSAFE_OBJECT_OPERATOR property_name | class_name T_PAAMAYIM_NEKUDOTAYIM simple_variable | new_variable T_PAAMAYIM_NEKUDOTAYIM simple_variable
- **static_member** → class_name T_PAAMAYIM_NEKUDOTAYIM simple_variable | variable_class_name T_PAAMAYIM_NEKUDOTAYIM simple_variable

### 11. Function Calls & Arguments
- **function_call** → name argument_list | T_READONLY argument_list | class_name T_PAAMAYIM_NEKUDOTAYIM member_name argument_list | variable_class_name T_PAAMAYIM_NEKUDOTAYIM member_name argument_list | callable_expr argument_list
- **argument_list** → '(' ')' | '(' non_empty_argument_list possible_comma ')' | '(' T_ELLIPSIS ')'
- **non_empty_argument_list** → argument | non_empty_argument_list ',' argument
- **argument** → expr | argument_no_expr
- **argument_no_expr** → identifier ':' expr | T_ELLIPSIS expr
- **clone_argument_list** → '(' ')' | '(' non_empty_clone_argument_list possible_comma ')' | '(' expr ',' ')' | '(' T_ELLIPSIS ')'
- **non_empty_clone_argument_list** → expr ',' argument | argument_no_expr | non_empty_clone_argument_list ',' argument
- **ctor_arguments** → ε | argument_list

### 12. Arrays
- **array_expr** → T_ARRAY '(' array_pair_list ')' | '[' array_pair_list ']'
- **array_pair_list** → non_empty_array_pair_list
- **non_empty_array_pair_list** → non_empty_array_pair_list ',' possible_array_pair | possible_array_pair
- **possible_array_pair** → ε | array_pair
- **array_pair** → expr T_DOUBLE_ARROW expr | expr | expr T_DOUBLE_ARROW ampersand variable | ampersand variable | T_ELLIPSIS expr | expr T_DOUBLE_ARROW T_LIST '(' array_pair_list ')' | T_LIST '(' array_pair_list ')'

### 13. String Literals & Interpolation
- **scalar** → T_LNUMBER | T_DNUMBER | T_START_HEREDOC T_ENCAPSED_AND_WHITESPACE T_END_HEREDOC | T_START_HEREDOC T_END_HEREDOC | T_START_HEREDOC encaps_list T_END_HEREDOC | dereferenceable_scalar | constant | class_constant
- **dereferenceable_scalar** → T_ARRAY '(' array_pair_list ')' | '[' array_pair_list ']' | T_CONSTANT_ENCAPSED_STRING | '"' encaps_list '"'
- **encaps_list** → encaps_list encaps_var | encaps_list T_ENCAPSED_AND_WHITESPACE | encaps_var | T_ENCAPSED_AND_WHITESPACE encaps_var
- **encaps_var** → T_VARIABLE | T_VARIABLE '[' encaps_var_offset ']' | T_VARIABLE T_OBJECT_OPERATOR T_STRING
- **backticks_expr** → ε | T_ENCAPSED_AND_WHITESPACE | encaps_list

### 14. Constants
- **constant** → name | T_LINE | T_FILE | T_DIR | T_TRAIT_C | T_METHOD_C | T_FUNC_C | T_PROPERTY_C | T_NS_C | T_CLASS_C
- **class_constant** → class_name T_PAAMAYIM_NEKUDOTAYIM identifier | variable_class_name T_PAAMAYIM_NEKUDOTAYIM identifier | class_name T_PAAMAYIM_NEKUDOTAYIM '{' expr '}' | variable_class_name T_PAAMAYIM_NEKUDOTAYIM '{' expr '}'

### 15. Namespace & Use Declarations
- **namespace_declaration** → T_NAMESPACE namespace_declaration_name ';' | T_NAMESPACE namespace_declaration_name '{' top_statement_list '}' | T_NAMESPACE '{' top_statement_list '}'
- **use_declarations** → use_declarations ',' use_declaration | use_declaration
- **use_declaration** → legacy_namespace_name | legacy_namespace_name T_AS T_STRING
- **group_use_declaration** → legacy_namespace_name T_NS_SEPARATOR '{' unprefixed_use_declarations possible_comma '}'
- **mixed_group_use_declaration** → legacy_namespace_name T_NS_SEPARATOR '{' inline_use_declarations possible_comma '}'
- **inline_use_declarations** → inline_use_declarations ',' inline_use_declaration | inline_use_declaration
- **inline_use_declaration** → unprefixed_use_declaration | use_type unprefixed_use_declaration
- **unprefixed_use_declarations** → unprefixed_use_declarations ',' unprefixed_use_declaration | unprefixed_use_declaration
- **unprefixed_use_declaration** → namespace_name | namespace_name T_AS T_STRING
- **use_type** → T_FUNCTION | T_CONST

### 16. Global & Static Variables
- **global_var_list** → global_var_list ',' global_var | global_var
- **global_var** → simple_variable
- **static_var_list** → static_var_list ',' static_var | static_var
- **static_var** → T_VARIABLE | T_VARIABLE '=' expr

### 17. Echo & Print
- **echo_expr_list** → echo_expr_list ',' echo_expr | echo_expr
- **echo_expr** → expr

### 18. For Loop Components
- **for_exprs** → ε | non_empty_for_exprs
- **for_cond_exprs** → ε | non_empty_for_exprs ',' expr | expr
- **non_empty_for_exprs** → non_empty_for_exprs ',' expr | non_empty_for_exprs ',' T_VOID_CAST expr | T_VOID_CAST expr | expr

### 19. Unset Variables
- **unset_variables** → unset_variable | unset_variables ',' unset_variable
- **unset_variable** → variable

### 20. Constant Declarations
- **const_list** → const_list ',' const_decl | const_decl
- **const_decl** → T_STRING '=' expr backup_doc_comment

### 21. Utility Rules
- **optional_expr** → ε | expr
- **possible_comma** → ε | ','
- **class_name** → T_STATIC | name
- **class_name_reference** → class_name | new_variable | '(' expr ')'
- **variable_class_name** → fully_dereferenceable
- **fully_dereferenceable** → variable | '(' expr ')' | dereferenceable_scalar | class_constant | new_dereferenceable
- **array_object_dereferenceable** → fully_dereferenceable | constant
- **callable_expr** → callable_variable | '(' expr ')' | dereferenceable_scalar | new_dereferenceable
- **member_name** → identifier | '{' expr '}' | simple_variable
- **property_name** → T_STRING | '{' expr '}' | simple_variable

### 22. Backup & Context Rules
- **backup_doc_comment** → ε (captures current doc comment)
- **backup_fn_flags** → ε (captures function flags)
- **backup_lex_pos** → ε (captures lexer position)
- **function** → T_FUNCTION (captures line number)
- **fn** → T_FN (captures line number)

## Implementation Priority Matrix

### ✅ COMPLETED (Current Implementation Status)
- [x] Basic expressions (arithmetic, logical, comparison)
- [x] Variables and assignments
- [x] Function declarations and calls
- [x] Class declarations (basic)
- [x] Control structures (if, while, for, foreach, switch)
- [x] Try-catch-finally blocks
- [x] Class methods with visibility
- [x] Class properties
- [x] Class constants
- [x] Namespace declarations
- [x] Use statements
- [x] Arrays (both syntaxes)
- [x] String literals and interpolation
- [x] Constants and magic constants

### 🔄 IN PROGRESS / PARTIAL
- [ ] **Attributes System (PHP 8.0+)**
  - Missing complete implementation of attribute parsing
  - Need attribute application to declarations
  
- [ ] **Property Hooks (PHP 8.4+)**
  - Basic structure exists but incomplete
  - Missing hook body parsing
  - Need get/set hook support

- [ ] **Union Types (PHP 8.0+)**
  - Basic union type parsing exists
  - Need complete intersection type support
  
- [ ] **Intersection Types (PHP 8.1+)**
  - Partial implementation
  - Need complex type expression support

### ❌ MISSING / TODO
- [ ] **Enum Declarations (PHP 8.1+)**
  - Complete enum syntax support
  - Backed enums with types
  - Enum cases with values

- [ ] **Match Expressions (PHP 8.0+)**
  - Full match syntax parsing
  - Multiple condition arms
  - Default arm support

- [ ] **Named Arguments (PHP 8.0+)**
  - Parameter name syntax
  - Named argument parsing in function calls

- [ ] **Arrow Functions (PHP 7.4+)**
  - fn keyword support
  - Short closure syntax
  - Auto-capture semantics

- [ ] **Nullsafe Operator (PHP 8.0+)**
  - ?-> operator parsing
  - Method call chains
  - Property access chains

- [ ] **Throw Expressions (PHP 8.0+)**
  - Throw as expression (not just statement)
  - Expression context support

- [ ] **Constructor Property Promotion (PHP 8.0+)**
  - Parameter visibility modifiers
  - Auto-property creation
  - Property hook integration

- [ ] **Readonly Properties/Classes (PHP 8.1+)**
  - Readonly modifier support
  - Class-level readonly
  - Property-level readonly

- [ ] **Anonymous Classes Enhanced**
  - Complete anonymous class support
  - Constructor arguments
  - Proper inheritance

- [ ] **Trait Adaptations**
  - Trait precedence rules
  - Method aliasing
  - Visibility modification

- [ ] **Advanced String Features**
  - Complex variable interpolation
  - Nowdoc improvements
  - Heredoc improvements

- [ ] **Pipe Operator (PHP 8.4+)**
  - |> operator parsing
  - Expression chaining

- [ ] **Alternative Syntax Enhancements**
  - Complete alternative control structure support
  - Mixed syntax detection

## Testing Strategy

### High Priority Test Cases Needed
1. **PHP 8.0+ Features**
   - Attributes on all declaration types
   - Union types in all contexts
   - Match expressions with complex conditions
   - Named arguments in function calls
   - Nullsafe operator chains

2. **PHP 8.1+ Features**
   - Enum declarations with backing types
   - Intersection types
   - Readonly properties and classes
   - First-class callable syntax

3. **PHP 8.4+ Features**
   - Property hooks (get/set)
   - Pipe operator expressions
   - Enhanced alternative syntax

4. **Edge Cases & Error Handling**
   - Malformed syntax recovery
   - Complex nested expressions
   - Mixed syntax patterns
   - Unicode identifier support

### Compatibility Testing
- Compare against PHP's official parser output
- Test with real-world PHP codebases
- Validate AST structure matches zend_ast.h
- Performance benchmarking for complex expressions

## Next Implementation Steps

### Phase 1: Complete PHP 8.0 Support (Priority: HIGH)
1. Implement attributes system completely
2. Add match expression parsing
3. Complete union type support
4. Add named arguments support
5. Implement nullsafe operator

### Phase 2: PHP 8.1+ Features (Priority: MEDIUM)
1. Add enum declaration support
2. Complete intersection type parsing
3. Implement readonly modifier support
4. Add first-class callable syntax

### Phase 3: PHP 8.4 Features (Priority: LOW)
1. Implement property hooks
2. Add pipe operator support
3. Enhance alternative syntax support

### Phase 4: Advanced Features (Priority: LOW)
1. Complete trait adaptation support
2. Enhanced anonymous classes
3. Complex string interpolation
4. Performance optimizations

---
*Generated from PHP Grammar Analysis of `/home/ubuntu/php-src/Zend/zend_language_parser.y` (lines 298-1674)*