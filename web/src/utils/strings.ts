// 字符串转蛇形格式
function toSnakeCase(str: string): string {
    return str.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`)
}

export function removeSuffix(originalString: string, suffixToRemove: string) {
    // 查找后缀在原字符串中的最后位置
    var index = originalString.lastIndexOf(suffixToRemove)

    // 如果找到了后缀，则截取前缀部分（不包含后缀）
    if (index !== -1) {
        return originalString.slice(0, index)
    }

    // 如果没找到后缀，直接返回原字符串
    return originalString
}

export default toSnakeCase
