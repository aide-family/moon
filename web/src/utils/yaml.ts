import yaml from 'js-yaml'

const yamlOptions: yaml.DumpOptions = {
    indent: 2,
    // 双引号
    quotingType: '"',
    // 字符串必须用双引号
    forceQuotes: true,
    schema: yaml.CORE_SCHEMA
}

const jsonOptions: yaml.LoadOptions = {
    schema: yaml.JSON_SCHEMA,
    json: true,
    onWarning: (e: any) => {
        throw e
    }
}

/**
 * 将json转换为yaml
 * @param data
 */
function toYaml(data: Object) {
    return yaml.dump(data, yamlOptions)
}

/**
 * 将yaml转换为json
 * @param data
 */
function toJson(data: string) {
    return yaml.load(data, jsonOptions)
}

export { toYaml, toJson }
