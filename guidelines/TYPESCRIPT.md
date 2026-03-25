---
description: 'Instructions for writing TypeScript code following the Google TypeScript Style Guide'
applyTo: '**/*.ts,**/*.tsx'
---

# TypeScript Development Instructions

Follow the [Google TypeScript Style Guide](https://google.github.io/styleguide/tsguide.html) when writing TypeScript code. These instructions highlight the core rules and idiomatic practices expected in this repository.

## General Rules

- Write simple, clear, and idiomatic TypeScript code.
- Always use `const` or `let` to declare variables. Use `const` by default. Never use `var`.
- Declare one variable per statement (e.g., avoid `let a = 1, b = 2;`).
- Use destructuring for objects and arrays where appropriate, but avoid deep destructuring that hinders readability.
- Prefer ES6 modules (`import`/`export`) over namespaces or internal modules.

## Naming Conventions

- **Variables & Functions**: Use `lowerCamelCase`.
- **Classes, Interfaces, Types, Enums**: Use `UpperCamelCase` (PascalCase).
- **Constants**: Use `CONSTANT_CASE` (all uppercase with underscores) for global, immutable values, or static readonly class properties.
- **Prefixes/Suffixes**: Do not use `_` as a prefix or suffix for identifiers (including private properties). `#` for native private fields is allowed.
- **File Names**: Use `lowerCamelCase` or `kebab-case` depending on the project structure (typically `kebab-case` in Angular projects, but match existing files).

## Type System

- **Prefer `interface`**: Use `interface` instead of `type` aliases for declaring shapes of objects, unless you specifically need union or intersection types.
- **Avoid `any`**: The use of `any` disables type checking. If the type is truly unknown upfront, use `unknown` and perform safe type narrowing.
- **Nullability**: Prefer using `undefined` over `null` to indicate the absence of a value, as it aligns better with JavaScript's default behavior and optional parameters.
- **Arrays**: Be consistent. `T[]` for simple types and `Array<T>` for more complex types or when expressing generic collections.
- **Type Inference**: Rely on TypeScript's inference for simple variables and return types when the type is obvious. Do not redundantly annotate types if the compiler infers them correctly (e.g., `const x: number = 5` is redundant).

## Classes and Objects

- Use arrow functions `() => {}` for callbacks to preserve the lexical scope of `this`.
- Always declare access modifiers (`public`, `private`, `protected`) for class members (except constructors if they don't have parameter properties).
- Use `readonly` for properties that should not be reassigned after initialization.
- Avoid the `this` keyword outside of classes, interfaces, or standard OOP paradigms.
- Prefer concise object literal syntax (e.g., `{a, b}` instead of `{a: a, b: b}`).

## Control Structures

- Prefer `for...of` loops over `for...in` when iterating over arrays or iterable objects. Use standard array methods (`.map`, `.filter`, `.reduce`) when they improve intent and readability.
- Return early to reduce nesting and prefer using `if { return }` over long if-else blocks.

## Formatting & Comments

- Use JSDoc format (`/** ... */`) for documenting classes, properties, and functions.
- Write comments that add information; avoid restating what the code obviously does.
- Place comments before decorators, not between decorators and the class/method itself.
- Do not use emoji in comments or code strings unless explicitly required by business logic.

## Summary Checklist

- [ ] Replaced `var` with `const` and `let`
- [ ] No `_` prefixes used in variable or property names
- [ ] Types are strict (no `any` without strong justification)
- [ ] `interface` preferred over `type`
- [ ] Callbacks use arrow functions
