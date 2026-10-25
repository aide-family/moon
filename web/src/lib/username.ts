/**
 * 用户名规则：只允许中英文和数字，长度最多 8 个字符。留空时由服务端从邮箱前缀派生，
 * 因此空值本身不算错误。
 * 必须与后端 app/alioth/internal/biz/user 的 usernamePattern 和
 * api/alioth/v1/identity.proto 上 RegisterRequest.name 的校验保持一致。
 */
export const USERNAME_PATTERN = /^[\p{L}\p{N}]{0,8}$/u

/** 用户名最大长度，供输入框与提示文案复用。 */
export const USERNAME_MAX_LENGTH = 8

/** 校验注册用户名，返回字段错误；留空或通过时返回 null。 */
export function usernameError(value: string, message: string): string | null {
  const name = value.trim()
  if (!name) {
    return null
  }
  return USERNAME_PATTERN.test(name) ? null : message
}
