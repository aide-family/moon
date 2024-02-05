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
        const storage = getValue(localValue, defaultValue)
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
        return () => {}
    }, [key])

    return [storedValue, setStorageValue, removeStorage]
}

function toString<T = string>(val: T): string {
    const t = typeof val
    let res = ''
    switch (t) {
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
        default:
            res = `${val}`
            break
    }

    return res
}

function getValue<T>(value: string | null, defalutVal: T): any {
    if (value === null) {
        return defalutVal
    }

    switch (typeof defalutVal) {
        case 'number':
            return +value
        case 'boolean':
            return value === 'true'
        case 'object':
            try {
                return JSON.parse(value)
            } catch (e) {
                return defalutVal
            }
        default:
            return value
    }
}

export default useStorage
