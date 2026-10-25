import feishu from "@/assets/brands/feishu.svg"

// 飞书品牌图标。直接复用 brands 下的 IconPark 授权素材，不再内联第二份路径，
// 与 GitHub/Gitee 图标一致，只通过 className 控制尺寸。
// 该 mark 自带品牌蓝 #3370FF：outline 按钮的文字色若套成 currentColor 会把 mark
// 压成灰色单色，失去品牌辨识度，因此这里保留素材原始配色。
export function FeishuIcon({ className }: { className?: string }) {
  return <img src={feishu} alt="" aria-hidden="true" className={className} />
}
