function randomString(length: number) {
    const chars =
        '0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ'
    let result = ''
    for (let i = length; i > 0; --i) {
        result += chars[Math.floor(Math.random() * chars.length)]
    }

    let hash = ''
    for (let i = 0; i < 32; i++) {
        hash += Math.floor(Math.random() * 16).toString(16)
    }
    return hash
}

const random = (min: number, max: number) => {
    return Math.floor(Math.random() * (max - min + 1) + min)
}

const defaultColors = [
    '#f56a00',
    '#7265e6',
    '#ffbf00',
    '#00a2ae',
    '#fb7293',
    '#00a854',
    '#f50',
    '#e6b600',
    '#6dc8ec',
    '#393d49'
]

const randomColor = () => {
    return defaultColors[random(0, defaultColors.length - 1)]
}

export {randomString, random, randomColor}
