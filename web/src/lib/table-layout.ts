// 表格列宽规范。
//
// 表格用「一列弹性 + 若干定宽列」：主内容列吃掉剩余宽度并允许换行，
// 短枚举列和操作列用固定宽度，避免不同行、不同页面之间列宽漂移。
//
// 这里的值必须是完整的 Tailwind 类名字面量，Tailwind 只会生成源码里出现过的类。

/** 各语义列的固定宽度。主内容列不设宽度，由表格剩余空间决定。 */
export const TABLE_COLUMN_WIDTH = {
  // 操作列：容纳两个内联操作（详情 / 更多），按钮本身用 nowrap 保持一行。
  actions: "w-[7rem]",
  // 状态、类型等短枚举徽标列。
  status: "w-[6.5rem]",
  type: "w-[7rem]",
  // 级别列：级别名通常 2–4 个字。
  level: "w-[6rem]",
  // 时间列：本地日期 + 时分秒。
  time: "w-[11rem]",
  // 负责人、管理员一类的身份列（姓名 + 邮箱，允许换行）。
  identity: "w-[12rem]",
  // UID、指纹一类的定长文本列。
  uid: "w-[15rem]",
} as const

/** 定宽列里的文本默认不换行；主内容列不要加这个类。 */
export const TABLE_NOWRAP = "whitespace-nowrap"

/**
 * 表格最小宽度分档。列越多、定宽列越多，就该用更大的档位；
 * 横向滚动由 Table / ListFrame 的内部滚动容器承担，不撑破页面。
 */
export const TABLE_MIN_WIDTH = {
  compact: "min-w-[56rem]",
  default: "min-w-[64rem]",
  wide: "min-w-[72rem]",
} as const

/**
 * CSS Grid 版表格的列模板（表头与每一行共用同一份）。
 * 左侧主列吃掉剩余宽度，其余列固定；单独写成字面量类名才能被 Tailwind 生成。
 */
export const TABLE_GRID_MAIN_IDENTITY_STATUS_ACTIONS =
  "md:grid-cols-[minmax(0,1fr)_12rem_6.5rem_7rem]"
