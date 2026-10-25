/** 图形验证码题型：NUMERIC 为随机 6 位数字码，MATH 为 100 以内四则运算。 */
export type CaptchaKind = "NUMERIC" | "MATH" | ""

export type Captcha = {
  id: string
  image: string
  kind: CaptchaKind
}
