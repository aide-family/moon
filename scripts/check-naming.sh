#!/usr/bin/env bash
# 命名检查：类型名只说明层次、构造函数不带类型名、接收者使用占位名时失败。
# 规则见 docs/iterations/conventions.md「命名」。
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

status=0
report() {
	echo "命名检查失败：$1" >&2
	shift
	printf '%s\n' "$@" >&2
	echo "规则见 docs/iterations/conventions.md「命名」。确认是误报时改检查规则，不要绕过检查。" >&2
	status=1
}

# 1) 禁止只说明层次的类型名。ent、conf、*.pb.go、wire_gen.go 是生成代码，不检查。
banned='^type (Usecase|UseCase|Biz|Manager|Helper|Util|Utils|Impl|Common|Base|Abstract|Repository|Handler|Controller|Data) struct'
hits=$(grep -rnE "$banned" app pkg api --include='*.go' \
	| grep -vE '/(ent|conf)/|\.pb\.go:|wire_gen\.go:' || true)
[ -z "$hits" ] || report "类型名只说明层次，应按业务对象或业务角色命名" "$hits"

# 2) 禁止 Usecase 写法回流。
hits=$(grep -rn 'Usecase' app pkg api --include='*.go' || true)
[ -z "$hits" ] || report "不要再用 Usecase 命名" "$hits"

# 3) 构造函数写 New<类型>。包名即被创建对象、且返回内置类型时允许 New（id.New）。
hits=$(grep -rnE '^func New\(' app pkg api --include='*.go' \
	| grep -vE '^app/alioth/pkg/id/id\.go:' || true)
[ -z "$hits" ] || report "构造函数要写 New<类型>" "$hits"

# 4) 接收者不使用占位名。
hits=$(grep -rnE '^func \((uc|this|self|me) \*' app pkg api --include='*.go' || true)
[ -z "$hits" ] || report "接收者要用类型名的业务小写形式" "$hits"

# 以下检查前端 web/src，规则见 docs/iterations/conventions.md「Web / 命名」。
if [ -d web/src ]; then
	# 5) 禁止通用占位文件名。
	hits=$(find web/src -type f \( -name '*.ts' -o -name '*.tsx' \) \
		| grep -E '/(utils|util|helpers|helper|common|misc|shared)\.tsx?$' || true)
	[ -z "$hits" ] || report "文件名要说明业务能力，不用 utils/helpers/common 这类占位名" "$hits"

	# 6) hooks 目录的文件名和导出要对上。
	for file in web/src/hooks/*.ts web/src/hooks/*.tsx; do
		[ -e "$file" ] || continue
		name=$(basename "$file")
		name=${name%.*}
		case "$name" in
		use-*) ;;
		*)
			report "hooks 文件名要写成 use-<hook 名>.ts" "$file"
			continue
			;;
		esac
		expected=$(printf '%s' "$name" | awk -F- '{out=$1; for (i=2; i<=NF; i++) out=out toupper(substr($i,1,1)) substr($i,2); print out}')
		if ! grep -qE "^export function $expected\b" "$file"; then
			report "hook 文件要和导出的 hook 同名（期望 $expected）" "$file"
		fi
	done

	# 7) 纯函数目录不放 hook，hook 放 hooks。
	hits=$(grep -rnE '^export function use[A-Z]' web/src/lib --include='*.ts' --include='*.tsx' || true)
	[ -z "$hits" ] || report "React hook 放 web/src/hooks，src/lib 只放纯函数" "$hits"
fi

exit "$status"
