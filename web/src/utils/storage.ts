// https://stackoverflow.com/questions/68424114/next-js-how-to-fetch-localstorage-data-before-client-side-rendering
// 解决 nextJS 无法获取初始localstorage问题

import { useEffect, useState } from 'react'

function useStorage<T = string>(
    key: string,
    defaultValue?: T
): [T | undefined, (value: T) => void, () => void] {
    const [storedValue, setStoredValue] = useState(defaultValue)

    const setStorageValue = (value: T) => {
        localStorage.setItem(key, toString(value))
        if (value !== storedValue) {
            setStoredValue(value)
        }
    }

    const removeStorage = () => {
        localStorage.removeItem(key)
    }

    useEffect(() => {
        const localValue = localStorage.getItem(key)
        const storage = getValue<T>(localValue || undefined) || defaultValue
        if (storage) {
            setStorageValue(storage)
            return
        }
        // 如果T是普通类型，则直接返回
        const t = typeof defaultValue

        switch (t) {
            case 'string':
                setStorageValue((localValue || defaultValue) as T)
                break
            case 'boolean':
                setStorageValue((localValue === 'true') as T)
                break
            case 'number':
                setStorageValue((Number(localValue) || defaultValue) as T)
                break
            default:
                setStorageValue(defaultValue as T)
        }
    }, [key])

    return [storedValue, setStorageValue, removeStorage]
}

function toString<T = string>(val: T): string {
    const t = typeof val
    let res = ''
    switch (t) {
        case 'string' || 'symbol' || 'boolean' || 'number' || 'bigint':
            res = `${val}`
            break
        case 'undefined' || 'null' || 'function':
            res = ''
            break
        case 'object' || 'array':
            try {
                res = JSON.stringify(val)
            } catch (e) {
                console.log('toString err', e)
            }
            break
    }

    return res
}

function getValue<T>(val?: string): T | undefined {
    if (val === undefined) return val
    try {
        // 尝试将输入字符串解析为 JSON 对象
        const parsedValue: T = JSON.parse(val)
        return parsedValue
    } catch (error) {
        // 如果解析失败，返回 undefined
        return undefined
    }
}

export default useStorage
