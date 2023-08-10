const colors: string[] = [
    "rgb(var(--green-3))",
    "rgb(var(--green-5))",
    "rgb(var(--green-7))",
    "rgb(var(--orange-4))",
    "rgb(var(--orange-6))",
    "rgb(var(--orange-7))",
    "rgb(var(--red-5))",
    "rgb(var(--red-6))",
    "rgb(var(--red-7))",
]

/**
 * 计算颜色
 * @param colors 颜色列表
 * @param val 当前值
 * @param max 最大值
 */
function calcColor(colors: string[], val: number, max: number): string {
    let colorIndex = 0

    if (val < max && val > 0) {
        colorIndex = Math.floor((val * colors.length) / max)
    }

    if (colorIndex < 0) {
        colorIndex = 0
    }

    if (val >= max || colorIndex >= colors.length) {
        colorIndex = colors.length - 1
    }

    return colors[colorIndex]
}

export {
    calcColor,
    colors
}