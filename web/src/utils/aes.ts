import CryptoJS from 'crypto-js'

const publicKey = '1234567890123456'
const iv = '1234567890123456'

/**
 * AES 加密
 * @param word: 需要加密的文本
 */

export const AES_Encrypt = (plaintext: string) => {
    const KEY = CryptoJS.enc.Utf8.parse(publicKey)
    const ciphertext = CryptoJS.AES.encrypt(plaintext, KEY, {
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7,
        iv: CryptoJS.enc.Utf8.parse(iv)
    }).toString()

    return ciphertext
}

/**
 * AES 解密
 * @param jsonStr
 */
export const AES_Decrypt = (jsonStr: string) => {
    const KEY = CryptoJS.enc.Utf8.parse(publicKey)
    const plaintext = CryptoJS.AES.decrypt(jsonStr, KEY, {
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7,
        iv: CryptoJS.enc.Utf8.parse(iv)
    }).toString(CryptoJS.enc.Utf8)

    return plaintext
}
