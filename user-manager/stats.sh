#!/bin/bash

# Script thống kê dự án
# Chạy: bash stats.sh hoặc ./stats.sh

echo "=========================================="
echo "📊 THỐNG KÊ DỰ ÁN"
echo "=========================================="

# Tìm git repository (có thể ở thư mục cha)
git_root=$(git rev-parse --show-toplevel 2>/dev/null)
if [ $? -ne 0 ]; then
    echo "❌ Không tìm thấy git repository!"
    exit 1
fi

echo "📁 Git repository: $(basename "$git_root")"
echo "📂 Thư mục hiện tại: $(basename "$(pwd)")"

echo

# 1. Tổng số commit
total_commits=$(git rev-list --count HEAD 2>/dev/null || echo "0")
echo "📈 Tổng số commit: $total_commits"

# 2. Tổng số file trong thư mục user-manager (loại trừ .git và các file ẩn)
total_files=$(find . -type f ! -path "./.git/*" ! -name ".*" | wc -l)
echo "📁 Tổng số file (user-manager): $total_files"

# 3. Tổng số dòng code trong thư mục user-manager
total_lines=$(find . -type f \( -name "*.go" -o -name "*.js" -o -name "*.ts" -o -name "*.py" -o -name "*.java" -o -name "*.c" -o -name "*.cpp" -o -name "*.h" -o -name "*.css" -o -name "*.html" -o -name "*.sql" -o -name "*.yml" -o -name "*.yaml" -o -name "*.json" \) ! -path "./.git/*" -exec wc -l {} + 2>/dev/null | tail -1 | awk '{print $1}' || echo "0")
echo "📝 Tổng số dòng code (user-manager): $total_lines"

# 4. Tổng số ngày code (từ commit đầu tiên đến hiện tại)
if [ "$total_commits" -gt 0 ]; then
    first_commit_date=$(git log --reverse --format="%ct" | head -1)
    current_date=$(date +%s)
    total_days=$(( (current_date - first_commit_date) / 86400 + 1 ))
    echo "📅 Tổng số ngày code: $total_days ngày"
    
    # Ngày bắt đầu và kết thúc
    start_date=$(date -d "@$first_commit_date" "+%d/%m/%Y")
    end_date=$(date "+%d/%m/%Y")
    echo "   Từ: $start_date đến $end_date"
else
    echo "📅 Tổng số ngày code: 0 ngày"
fi

echo

# 5. Thống kê commit theo ngày (10 ngày gần nhất)
echo "📊 THỐNG KÊ COMMIT THEO NGÀY (10 ngày gần nhất):"
echo "------------------------------------------"

if [ "$total_commits" -gt 0 ]; then
    # Lấy commit trong 10 ngày gần nhất
    git log --since="10 days ago" --format="%cd" --date=short | sort | uniq -c | sort -rn | head -10 | while read count date; do
        # Tạo thanh progress đơn giản
        bar=""
        for ((i=1; i<=count; i++)); do
            if [ $i -le 20 ]; then  # Giới hạn 20 ký tự
                bar="${bar}█"
            fi
        done
        [ ${#bar} -eq 0 ] && bar="▌"
        printf "%-12s %3d commits %s\n" "$date" "$count" "$bar"
    done
    
    echo
    echo "📈 TỔNG QUAN:"
    echo "------------------------------------------"
    
    # Commit nhiều nhất trong 1 ngày
    max_commits_day=$(git log --format="%cd" --date=short | sort | uniq -c | sort -rn | head -1)
    if [ -n "$max_commits_day" ]; then
        echo "🔥 Ngày nhiều commit nhất: $max_commits_day"
    fi
    
    # Trung bình commit/ngày
    if [ "$total_days" -gt 0 ]; then
        avg_commits=$(echo "scale=2; $total_commits / $total_days" | bc 2>/dev/null || echo "0")
        echo "📊 Trung bình: $avg_commits commits/ngày"
    fi
    
    # Tác giả chính
    main_author=$(git log --format="%an" | sort | uniq -c | sort -rn | head -1 | sed 's/^ *[0-9]* *//')
    echo "👤 Tác giả chính: $main_author"
    
else
    echo "Chưa có commit nào!"
fi

echo
echo "=========================================="
echo "✅ Hoàn thành thống kê!"
echo "=========================================="
