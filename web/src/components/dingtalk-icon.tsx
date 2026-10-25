import dingtalk from "@/assets/brands/dingtalk.svg"

// 钉钉品牌图标。复用 brands 下的素材，和飞书图标一样不再内联第二份路径，
// 只通过 className 控制尺寸。该 mark 自带品牌蓝 #0089FF：套成 currentColor
// 会把 mark 压成灰色单色，失去品牌辨识度，因此保留素材原始配色。
export function DingtalkIcon({ className }: { className?: string }) {
  return (
    <img src={dingtalk} alt="" aria-hidden="true" className={className} />
  )
}
