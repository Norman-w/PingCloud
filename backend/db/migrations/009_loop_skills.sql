-- Rename 正手前冲弧圈 → 正手拉反胶防守球
UPDATE skills SET name='正手拉反胶防守球' WHERE id=3;

-- Add new 拉球 skills
INSERT INTO skills (category, name, style, sort_order) VALUES
('基本功', '正手反胶对拉', '双反', 3),
('基本功', '正手拉前冲', '双反', 4),
('基本功', '正手拉侧弧圈', '双反', 5),
('基本功', '正手拉长胶削球', '双反', 6)
ON CONFLICT (category, name, style) DO NOTHING;
