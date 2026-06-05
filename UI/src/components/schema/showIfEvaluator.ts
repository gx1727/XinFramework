import type { ShowIfCondition } from "@/types/schema"

export function evaluateShowIf(
  condition: ShowIfCondition,
  formValues: Record<string, unknown>
): boolean {
  if (!condition) return true

  if (condition.dependsOn) {
    const fieldValue = formValues[condition.dependsOn]

    if (condition.equals !== undefined) {
      if (fieldValue !== condition.equals) return false
    }

    if (condition.notEquals !== undefined) {
      if (fieldValue === condition.notEquals) return false
    }

    if (condition.in !== undefined) {
      if (!Array.isArray(condition.in)) return false
      if (!condition.in.includes(fieldValue)) return false
    }

    if (condition.notIn !== undefined) {
      if (!Array.isArray(condition.notIn)) return false
      if (condition.notIn.includes(fieldValue)) return false
    }

    if (condition.contains !== undefined) {
      if (Array.isArray(fieldValue) && !fieldValue.includes(condition.contains)) {
        return false
      }
      if (typeof fieldValue === "string" && !fieldValue.includes(String(condition.contains))) {
        return false
      }
    }

    if (condition.greaterThan !== undefined) {
      if (typeof fieldValue !== "number" || fieldValue <= condition.greaterThan) {
        return false
      }
    }

    if (condition.lessThan !== undefined) {
      if (typeof fieldValue !== "number" || fieldValue >= condition.lessThan) {
        return false
      }
    }
  }

  if (condition.and && condition.and.length > 0) {
    for (const subCondition of condition.and) {
      if (!evaluateShowIf(subCondition, formValues)) {
        return false
      }
    }
  }

  if (condition.or && condition.or.length > 0) {
    let orResult = false
    for (const subCondition of condition.or) {
      if (evaluateShowIf(subCondition, formValues)) {
        orResult = true
        break
      }
    }
    if (!orResult) return false
  }

  return true
}

export function buildShowIfFromString(expression: string): ShowIfCondition | null {
  const match = expression.match(/^(\w+)\s*(===|!==|>=|<=|>|<|in)\s*(.+)$/)
  if (!match) return null

  const [, field, operator, valueStr] = match
  const value = parseValue(valueStr)

  switch (operator) {
    case "===":
      return { dependsOn: field, equals: value }
    case "!==":
      return { dependsOn: field, notEquals: value }
    case ">":
      return { dependsOn: field, greaterThan: Number(value) }
    case "<":
      return { dependsOn: field, lessThan: Number(value) }
    case ">=":
      return { dependsOn: field, greaterThan: Number(value) - 1 }
    case "<=":
      return { dependsOn: field, lessThan: Number(value) + 1 }
    case "in":
      return { dependsOn: field, in: Array.isArray(value) ? value : [value] }
    default:
      return null
  }
}

function parseValue(valueStr: string): unknown {
  const trimmed = valueStr.trim()
  
  if (trimmed === "true") return true
  if (trimmed === "false") return false
  if (trimmed === "null") return null
  if (trimmed === "undefined") return undefined
  
  if (trimmed.startsWith("[") && trimmed.endsWith("]")) {
    try {
      return JSON.parse(trimmed)
    } catch {
      return trimmed
    }
  }
  
  if (trimmed.startsWith("'") && trimmed.endsWith("'")) {
    return trimmed.slice(1, -1)
  }
  if (trimmed.startsWith('"') && trimmed.endsWith('"')) {
    return trimmed.slice(1, -1)
  }
  
  const num = Number(trimmed)
  if (!isNaN(num)) return num
  
  return trimmed
}
