package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"nonlegacy-garment-3d/server/internal/config"
	"nonlegacy-garment-3d/server/internal/database"
	"nonlegacy-garment-3d/server/internal/model"
	"nonlegacy-garment-3d/server/internal/router"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	seed := flag.Bool("seed", false, "插入种子数据")
	flag.Parse()

	// 加载配置
	cfg := config.Load()

	// 确保目录存在
	os.MkdirAll(filepath.Dir(cfg.DBPath), 0755)
	os.MkdirAll(cfg.UploadDir, 0755)
	os.MkdirAll("reports", 0755)

	// 初始化数据库
	db, err := database.NewSQLiteDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer db.Close()

	// 运行迁移
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化 Store
	garmentStore := model.NewGarmentStore(db)
	modelStore := model.NewModelStore(db)
	userStore := model.NewUserStore(db)
	favoriteStore := model.NewFavoriteStore(db)
	dailyPickStore := model.NewDailyPickStore(db)
	submissionStore := model.NewSubmissionStore(db)
	commentStore := model.NewCommentStore(db)
	annotationStore := model.NewAnnotationStore(db)
	quizStore := model.NewQuizStore(db)

	// 插入种子数据
	if *seed {
		seedData(db, userStore, garmentStore, modelStore, dailyPickStore, cfg)
	}

	// 设置路由
	r := router.SetupRouter(db, garmentStore, modelStore, userStore, favoriteStore, dailyPickStore, submissionStore, commentStore, annotationStore, quizStore, cfg)

	// 启动服务器
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("[Server] 非遗华裳 API 启动于 http://localhost%s", addr)
	log.Printf("[Server] 环境: %s", cfg.AppEnv)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

func seedData(
	db *sql.DB,
	userStore *model.UserStore,
	garmentStore *model.GarmentStore,
	modelStore *model.ModelStore,
	dailyPickStore *model.DailyPickStore,
	cfg *config.Config,
) {
	log.Println("[Seed] 开始插入种子数据...")

	// 清理已有数据，支持重复运行 -seed（保留 users 表）
	db.Exec("DELETE FROM daily_picks")
	db.Exec("DELETE FROM favorites")
	db.Exec("DELETE FROM models")
	db.Exec("DELETE FROM garments")
	// 模型行已清空，同步清掉旧的模型文件目录（保留 submissions 图片目录），
	// 避免重复 -seed 时按新自增 ID 不断生成孤儿目录
	if entries, err := os.ReadDir(cfg.UploadDir); err == nil {
		for _, e := range entries {
			if e.IsDir() && e.Name() != "submissions" {
				os.RemoveAll(filepath.Join(cfg.UploadDir, e.Name()))
			}
		}
	}
	log.Println("[Seed] 已清理旧种子数据")

	// 创建管理员用户
	adminHash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin, err := userStore.Create("admin", "admin@nonlegacy.cn", string(adminHash))
	if err == nil {
		// 手动更新 role 为 admin
		db.Exec("UPDATE users SET role = 'admin' WHERE id = ?", admin.ID)
		log.Printf("[Seed] 管理员用户创建成功 (ID: %d, 密码: admin123)", admin.ID)
	} else {
		log.Printf("[Seed] 管理员用户可能已存在: %v", err)
		// 确保已有 admin 用户也具有管理员角色
		db.Exec("UPDATE users SET role = 'admin' WHERE username = 'admin'")
	}

	// 创建普通用户
	userHash, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
	if _, err := userStore.Create("user", "user@nonlegacy.cn", string(userHash)); err == nil {
		log.Println("[Seed] 普通用户创建成功 (用户名: user, 密码: user123)")
	}

	// 创建种子服饰 — 35 件，覆盖 22 个省份，4 个密度层级
	garments := []struct {
		Name            string
		Dynasty         string
		Ethnic          string
		Craft           string
		Material        string
		Pattern         string
		CulturalMeaning string
		Province        string
		Description     string
		ModelFile       string
	}{
		// ═══════════════ 高密度：贵州 (5件) ═══════════════
		{
			Name: "苗族百褶裙", Dynasty: "清", Ethnic: "苗族", Craft: "蜡染+刺绣",
			Material: "棉麻", Pattern: "几何纹+花卉纹", CulturalMeaning: "象征苗族人民对自然和生命的崇敬，百褶代表山川河流的绵延不绝",
			Province: "贵州", Description: "苗族百褶裙是苗族女性传统服饰中最具代表性的下装，以精细的蜡染工艺和繁复的刺绣图案著称。每条裙子需经过数十道工序制作完成，裙摆多达数百褶皱。",
			ModelFile: "苗族百褶裙.glb",
		},
		{
			Name: "苗族银衣", Dynasty: "清", Ethnic: "苗族", Craft: "银饰锻造+刺绣",
			Material: "棉布+银", Pattern: "龙凤纹+花卉纹", CulturalMeaning: "银饰象征着财富与吉祥，也是苗族人民驱邪避凶的护身符，代代相传的银匠技艺承载着苗族的民族记忆",
			Province: "贵州", Description: "苗族银衣是苗族盛装中最华丽的服饰，全身缀满手工打制的银片、银铃和银角，单件重量可达十余公斤，是苗族银饰锻造技艺的集大成之作。",
		},
		{
			Name: "侗族亮布衣", Dynasty: "清", Ethnic: "侗族", Craft: "亮布制作+刺绣",
			Material: "棉布", Pattern: "太阳纹+几何纹", CulturalMeaning: "亮布在侗族文化中象征光明与希望，制作过程中反复捶打使布面呈现金属般光泽，体现侗族人民对美的执着追求",
			Province: "贵州", Description: "侗族亮布衣选用自织棉布，经蓝靛多次浸染后用木槌反复捶打至上万次，使布面呈现紫铜色的金属光泽。配以精致的刺绣点缀，是侗族女性节庆时的盛装。",
		},
		{
			Name: "布依族蜡染裙", Dynasty: "清", Ethnic: "布依族", Craft: "蜡染",
			Material: "棉麻", Pattern: "漩涡纹+鱼纹", CulturalMeaning: "蜡染中的漩涡纹象征布依族对水的崇拜，鱼纹代表丰收与富足，蓝白相间的图案是布依族人与自然和谐共生的写照",
			Province: "贵州", Description: "布依族蜡染裙以铜刀蘸熔蜡在布面绘制图案，经蓝靛浸染后脱蜡显花，蓝底白花的纹样清雅古朴。布依族女性从小学习蜡染技艺，每条裙子都是独一无二的艺术品。",
		},
		{
			Name: "苗族挑花头帕", Dynasty: "明", Ethnic: "苗族", Craft: "挑花",
			Material: "棉布", Pattern: "八角花纹+蝴蝶纹", CulturalMeaning: "蝴蝶是苗族创世神话中万物之母，八角花象征宇宙八方，挑花技法以布面经纬为坐标，体现苗族独特的空间观念",
			Province: "贵州", Description: "苗族挑花头帕采用数纱挑花技法，在平纹布上按经纬线挑绣出对称的几何图案，无需画稿，全凭心算。图案精细规整，是苗族女性智慧的结晶。",
		},

		// ═══════════════ 高密度：云南 (4件) ═══════════════
		{
			Name: "傣族筒裙", Dynasty: "清", Ethnic: "傣族", Craft: "织锦",
			Material: "棉", Pattern: "几何纹+孔雀纹", CulturalMeaning: "孔雀是傣族的吉祥物，织锦工艺代代相传，体现了傣族女性的勤劳与智慧",
			Province: "云南", Description: "傣族筒裙以紧身合体的设计和绚丽多彩的织锦图案为特色，是傣族女性日常穿着和节日盛装的重要组成部分。",
			ModelFile: "傣族筒裙.glb",
		},
		{
			Name: "彝族蜡染百褶裙", Dynasty: "清", Ethnic: "彝族", Craft: "蜡染+贴布绣",
			Material: "棉麻", Pattern: "火焰纹+羊角纹", CulturalMeaning: "火焰纹象征彝族对火的崇拜与敬畏，羊角纹代表财富与地位，百褶设计使行走时裙摆如火焰般摇曳",
			Province: "云南", Description: "彝族蜡染百褶裙融合蜡染与贴布绣双重工艺，裙身以蜡染绘制火焰图案，裙摆用彩色布片贴绣几何纹样。百褶设计让裙摆展开时如孔雀开屏般绚丽。",
			ModelFile: "彝族.glb",
		},
		{
			Name: "白族扎染上衣", Dynasty: "现代", Ethnic: "白族", Craft: "扎染",
			Material: "棉布", Pattern: "蝴蝶花纹+几何纹", CulturalMeaning: "扎染的每一个结都蕴含着白族妇女对生活的美好祝愿，蓝白相间的色彩搭配反映了白族崇尚清新素雅的审美情趣",
			Province: "云南", Description: "白族扎染上衣采用大理周城传承千年的扎染技艺，通过纱线捆扎布料后浸入板蓝根染液，解开后形成自然晕染的蓝白花纹。每一件都是独一无二的艺术品。",
		},
		{
			Name: "哈尼族树皮衣", Dynasty: "古代", Ethnic: "哈尼族", Craft: "树皮加工",
			Material: "树皮纤维", Pattern: "素面+编织纹", CulturalMeaning: "树皮衣是人类最古老的服饰形态之一，哈尼族保留这一传统是对祖先生活方式的活态传承，体现了人与自然最原始的和谐",
			Province: "云南", Description: "哈尼族树皮衣选用构树等树皮，经浸泡、捶打、漂洗等多道工序制成柔软的树皮纤维布，再缝制成衣。这是人类服饰文明的活化石，已被列为非物质文化遗产。",
		},

		// ═══════════════ 中高密度：四川 (3件) ═══════════════
		{
			Name: "彝族擦尔瓦", Dynasty: "清", Ethnic: "彝族", Craft: "羊毛擀毡+挑花",
			Material: "羊毛", Pattern: "几何条纹+日月纹", CulturalMeaning: "擦尔瓦是彝族人民的身份象征，羊毛的白色代表纯洁与高贵，披毡的形制体现了彝族适应高寒山区的生存智慧",
			Province: "四川", Description: "彝族擦尔瓦为羊毛擀制的披毡，质地厚实保暖，边缘饰有精美的挑花图案。白天为衣，夜间为被，是彝族人民在高寒山区生活中的必备之物。",
		},
		{
			Name: "羌族云云鞋", Dynasty: "现代", Ethnic: "羌族", Craft: "挑花刺绣",
			Material: "棉布+皮底", Pattern: "云纹+羊角纹", CulturalMeaning: "云云鞋上的云纹是羌族人对天的敬仰，羊是羌族的图腾，鞋尖上翘如羊角，承载着羌族从游牧到农耕的文化记忆",
			Province: "四川", Description: "羌族云云鞋因鞋面上绣有云纹图案而得名，鞋尖微微上翘，造型独特。鞋帮以精细的挑花和刺绣工艺装饰，色彩艳丽，是羌族服饰文化的重要标志。",
		},
		{
			Name: "四川蜀绣嫁衣", Dynasty: "清", Ethnic: "汉族", Craft: "蜀绣",
			Material: "丝绸", Pattern: "龙凤呈祥+芙蓉花", CulturalMeaning: "蜀绣嫁衣承载着川蜀女子对美满婚姻的期盼，芙蓉花是成都的象征，龙凤纹代表夫妻和谐美满",
			Province: "四川", Description: "四川蜀绣嫁衣以蜀绣为核心工艺，运用晕针、切针、拉针等多种蜀绣独特针法，将龙凤和芙蓉花绣于大红绸缎之上。绣工精细，色彩丰富，是中国四大名绣之一的典型代表。",
		},

		// ═══════════════ 中高密度：西藏 (3件) ═══════════════
		{
			Name: "藏族氆氇袍", Dynasty: "清", Ethnic: "藏族", Craft: "氆氇织造",
			Material: "羊毛", Pattern: "几何条纹", CulturalMeaning: "反映藏族人民对高原生活的适应与智慧，氆氇是藏族最具代表性的手工织品",
			Province: "西藏", Description: "藏族氆氇袍采用传统手工织造的氆氇面料制成，厚实保暖，适应高原多变的气候条件。袍身宽大，便于活动。",
			ModelFile: "藏族氆氇袍.glb",
		},
		{
			Name: "藏族邦典围裙", Dynasty: "现代", Ethnic: "藏族", Craft: "氆氇织造",
			Material: "羊毛+丝线", Pattern: "彩虹条纹", CulturalMeaning: "邦典围裙的彩色条纹象征彩虹，是藏族女性成年的标志，每一种颜色代表一种美好的祝福",
			Province: "西藏", Description: "邦典围裙是藏族已婚女性的标志性服饰，以彩色羊毛和丝线编织成横向彩虹条纹，色彩明快。日喀则地区的邦典织造技艺尤为精湛，已被列入国家级非物质文化遗产。",
		},
		{
			Name: "门巴族氆氇袍", Dynasty: "清", Ethnic: "门巴族", Craft: "氆氇织造",
			Material: "羊毛", Pattern: "条纹+十字纹", CulturalMeaning: "门巴族氆氇袍上的十字纹象征太阳和四方，体现门巴族原始苯教信仰中对自然的崇拜",
			Province: "西藏", Description: "门巴族氆氇袍与藏族氆氇袍一脉相承又独具特色，袍身以黑色或深红色为底，配以白色十字纹装饰。门巴族世代生活在喜马拉雅山南麓，氆氇袍是他们适应高寒环境的重要发明。",
		},

		// ═══════════════ 中高密度：新疆 (3件) ═══════════════
		{
			Name: "维吾尔艾德莱斯裙", Dynasty: "清", Ethnic: "维吾尔族", Craft: "艾德莱斯绸织造",
			Material: "丝绸", Pattern: "巴旦木纹+梳子纹+波浪纹", CulturalMeaning: "艾德莱斯绸的图案源于维吾尔族对自然的观察和抽象提炼，流动的色彩如沙漠中的绿洲般珍贵，代表了维吾尔族人民的热情与创造力",
			Province: "新疆", Description: "维吾尔艾德莱斯裙采用和田、喀什地区传统扎染经线织造工艺制成。丝绸经线按图案分段扎染后再织造，产生如水彩晕染般的独特效果，每件都是无法复制的孤品。",
			ModelFile: "维吾尔艾德莱斯裙.glb",
		},
		{
			Name: "哈萨克族绣花大衣", Dynasty: "清", Ethnic: "哈萨克族", Craft: "刺绣+贴花",
			Material: "羊毛呢+丝绸", Pattern: "角纹+植物纹+鹰纹", CulturalMeaning: "鹰是哈萨克族的精神图腾，角纹象征牛羊成群、生活富足，绣花大衣承载着草原民族的审美理想和生命活力",
			Province: "新疆", Description: "哈萨克族绣花大衣选用厚实的羊毛呢料，在领口、袖口和下摆处施以精美的刺绣和贴花工艺。纹样以羊角纹和植物纹为主，色彩明艳对比强烈，展现了草原民族的豪迈与热情。",
		},
		{
			Name: "柯尔克孜族白毡帽", Dynasty: "清", Ethnic: "柯尔克孜族", Craft: "擀毡+刺绣",
			Material: "羊毛毡", Pattern: "几何纹+卷草纹", CulturalMeaning: "白毡帽是柯尔克孜族男子的标志，白色象征雪山般纯洁的品格，帽檐上翻的造型如天山雪峰般巍峨",
			Province: "新疆", Description: "柯尔克孜族白毡帽由优质白羊毛擀制而成，帽顶和帽檐绣有精美的几何图案。帽檐两侧微微上翻，造型独特，不仅御寒保暖，更是柯尔克孜族民族认同的重要标志。",
		},

		// ═══════════════ 中密度：广西 (2件) ═══════════════
		{
			Name: "壮族织锦被面", Dynasty: "清", Ethnic: "壮族", Craft: "壮锦织造",
			Material: "棉+丝线", Pattern: "菱形纹+蛙纹+太阳纹", CulturalMeaning: "蛙是壮族的图腾，象征雨水与丰收，太阳纹代表光明与生命，壮锦的每一根丝线都编织着壮族对美好生活的向往",
			Province: "广西", Description: "壮族织锦被面是壮锦工艺的代表作，以棉线为经、彩色丝线为纬，采用通经断纬技法织出精美的几何图案。色彩对比强烈，结构严谨，与云锦、蜀锦、宋锦并称中国四大名锦。",
			ModelFile: "壮族织锦被面.glb",
		},
		{
			Name: "瑶族刺绣盘王衣", Dynasty: "清", Ethnic: "瑶族", Craft: "刺绣+挑花",
			Material: "棉布", Pattern: "盘王印+龙犬纹", CulturalMeaning: "盘王是瑶族的始祖，盘王印图案是瑶族子孙认祖归宗的标志，承载着瑶族迁徙的历史记忆和民族认同",
			Province: "广西", Description: "瑶族盘王衣是瑶族祭祀盘王时穿着的盛装，通身绣满盘王印、龙犬纹等传统图案。绣工精细繁复，一件盘王衣往往需要数年时间才能完成，是瑶族刺绣技艺的最高体现。",
		},

		// ═══════════════ 中密度：湖南 (2件) ═══════════════
		{
			Name: "土家族织锦衣裙", Dynasty: "清", Ethnic: "土家族", Craft: "土家织锦",
			Material: "棉+丝线", Pattern: "台台花纹+阳雀花纹", CulturalMeaning: "台台花是土家族对虎的抽象表达，体现了白虎图腾崇拜；阳雀花象征报春与希望，寄托着土家族对大自然的深情",
			Province: "湖南", Description: "土家族织锦衣裙以土家织锦（西兰卡普）工艺制成，采用古老的斜织腰机，通经断纬，纬线起花。图案多达一百余种，色彩古朴厚重，被誉为纺织业的活化石。",
		},
		{
			Name: "苗族花带裙", Dynasty: "明", Ethnic: "苗族", Craft: "织带+刺绣",
			Material: "棉+丝线", Pattern: "菱形纹+八角花纹", CulturalMeaning: "苗族花带是苗族女性传递情感的媒介，每一条花带的图案都蕴含着编织者的祝福和心意",
			Province: "湖南", Description: "苗族花带裙以数十条手工编织的花带拼接而成，每条花带宽仅数厘米却织有精美繁复的几何图案。摇摆时花带飘动，色彩斑斓，是湘西苗族女性最具特色的服饰之一。",
		},

		// ═══════════════ 中密度：内蒙古 (2件) ═══════════════
		{
			Name: "蒙古族长袍", Dynasty: "元", Ethnic: "蒙古族", Craft: "镶边刺绣",
			Material: "绸缎", Pattern: "云纹+动物纹", CulturalMeaning: "象征蒙古族对草原和自然的敬畏，展现了游牧民族的豪迈气概",
			Province: "内蒙古", Description: "蒙古族长袍以其宽大的版型和精美的镶边工艺著称，适应草原游牧生活的需要，兼具实用性与艺术性。",
			ModelFile: "蒙古族长袍.glb",
		},
		{
			Name: "蒙古族喀尔喀袍", Dynasty: "清", Ethnic: "蒙古族", Craft: "镶边+贴花刺绣",
			Material: "绸缎", Pattern: "如意纹+盘肠纹", CulturalMeaning: "如意纹和盘肠纹是蒙古族对吉祥如意的祈愿，袍服的每一道镶边都代表着一重祝福",
			Province: "内蒙古", Description: "蒙古族喀尔喀袍是喀尔喀蒙古部的传统礼服，以多层镶边和精美的贴花刺绣为特色。肩部隆起如山峰，象征蒙古族对圣山的崇拜，整体造型端庄大气。",
			ModelFile: "蒙古族喀尔喀袍.glb",
		},

		// ═══════════════ 中密度：江苏 (2件) ═══════════════
		{
			Name: "苏绣旗袍", Dynasty: "现代", Ethnic: "汉族", Craft: "苏绣",
			Material: "丝绸", Pattern: "花鸟纹", CulturalMeaning: "体现江南水乡的柔美与精致，展现中国传统女性的优雅气质",
			Province: "江苏", Description: "苏绣旗袍融合了苏州刺绣的精湛技艺与传统旗袍的优雅线条，以细腻的针法和丰富的色彩层次闻名于世。",
			ModelFile: "苏绣旗袍.glb",
		},
		{
			Name: "汉族缂丝吉服", Dynasty: "明", Ethnic: "汉族", Craft: "缂丝",
			Material: "丝绸+金线", Pattern: "花卉纹+云龙纹", CulturalMeaning: "缂丝有一寸缂丝一寸金之说，是丝织工艺的巅峰。吉服上的纹样承载着对穿着者的最高祝福与尊崇",
			Province: "江苏", Description: "汉族缂丝吉服采用苏州缂丝技艺，以通经断纬的方法织造，图案与面料一体成型。缂丝技艺极其繁复，一件吉服往往需要数年才能织成，是皇家贵族专属的奢华服饰。",
		},

		// ═══════════════ 中密度：浙江 (2件) ═══════════════
		{
			Name: "汉族杭绣礼服", Dynasty: "现代", Ethnic: "汉族", Craft: "杭绣",
			Material: "丝绸", Pattern: "西湖景致+花卉纹", CulturalMeaning: "杭绣以西湖山水为灵感源泉，将江南的诗情画意融入一针一线，体现了杭州作为丝绸之府的深厚文化底蕴",
			Province: "浙江", Description: "汉族杭绣礼服以杭州传统刺绣为基础，针法细腻多变，色彩清雅秀丽。杭绣传承南宋宫廷刺绣遗风，与苏绣、湘绣、蜀绣、粤绣齐名，是江南刺绣的重要流派。",
		},
		{
			Name: "畲族凤凰装", Dynasty: "清", Ethnic: "畲族", Craft: "刺绣+织带",
			Material: "棉布+丝线", Pattern: "凤凰纹+牡丹纹", CulturalMeaning: "凤凰装是畲族女性的婚礼盛装，凤凰是畲族的图腾，传说畲族始祖是凤凰的化身，穿上凤凰装即获得凤凰的祝福",
			Province: "浙江", Description: "畲族凤凰装以红色为主调，头戴凤凰冠，身穿绣有凤凰图案的上衣，腰系织带。整套服饰色彩艳丽，展现了畲族独特的审美追求和民族自豪感。",
		},

		// ═══════════════ 低密度：各1件 (12省) ═══════════════
		{
			Name: "广东潮绣婚服", Dynasty: "清", Ethnic: "汉族", Craft: "潮绣",
			Material: "丝绸+金线", Pattern: "龙凤呈祥+牡丹纹", CulturalMeaning: "潮绣婚服以金线盘绣的立体浮雕效果象征婚姻的珍贵与永恒，是潮汕地区婚俗文化的重要载体",
			Province: "广东", Description: "广东潮绣婚服以潮绣特有的垫高绣法著称，使用金线和银线盘绣出龙凤、牡丹等吉祥图案，绣面具有强烈的立体浮雕感。潮绣是中国四大名绣粤绣的重要分支，以富丽堂皇的风格闻名。",
			ModelFile: "广东潮绣婚服.glb",
		},
		{
			Name: "福建惠安女服饰", Dynasty: "现代", Ethnic: "汉族", Craft: "刺绣+拼接",
			Material: "棉布", Pattern: "几何纹+花卉纹", CulturalMeaning: "惠安女服饰的「封建头、民主肚、节约衣、浪费裤」造型体现了闽南女性勤劳坚韧、开放包容的独特品格",
			Province: "福建", Description: "福建惠安女服饰造型独特：黄斗笠、花头巾、短上衣、宽腿裤。上衣紧身短小露出肚脐，裤腿却异常宽大，配色大胆鲜艳，是中国最具辨识度的地域性传统服饰之一。",
			ModelFile: "福建惠安女服饰.glb",
		},
		{
			Name: "藏族安多袍", Dynasty: "清", Ethnic: "藏族", Craft: "氆氇+羔皮镶边",
			Material: "羊毛+羔皮", Pattern: "纯色+皮饰", CulturalMeaning: "安多袍的羔皮镶边是身份的象征，袍身的厚重体现了高原人民对自然的敬畏与适应",
			Province: "青海", Description: "藏族安多袍是安多藏区的传统袍服，以厚实的氆氇为面料，领口、袖口和下摆镶有名贵的羔皮。袍身宽大，束腰后形成自然的兜囊，可携带随身物品，是游牧生活的智慧体现。",
			ModelFile: "藏族安多袍.glb",
		},
		{
			Name: "裕固族红缨帽", Dynasty: "清", Ethnic: "裕固族", Craft: "刺绣+珠绣",
			Material: "羊毛毡+珊瑚珠", Pattern: "植物纹+几何纹", CulturalMeaning: "红缨帽上的红缨象征太阳，珊瑚珠代表珍贵与吉祥，帽子造型反映了裕固族从西域迁徙而来的历史渊源",
			Province: "甘肃", Description: "裕固族红缨帽是裕固族女性最具特色的头饰，尖顶帽身以白羊毛擀制，帽顶缀有红色流苏，帽身绣满彩色丝线和珊瑚珠。裕固族是中国人口较少的民族之一，其服饰文化弥足珍贵。",
			ModelFile: "裕固族红缨帽.glb",
		},
		{
			Name: "回族绣花盖头", Dynasty: "现代", Ethnic: "回族", Craft: "刺绣",
			Material: "丝绸+纱", Pattern: "花卉纹+几何纹+阿拉伯文纹", CulturalMeaning: "盖头是回族女性信仰与美的结合，绣花图案融合了中国传统纹样与伊斯兰文化元素，体现了两大文明的和谐交融",
			Province: "宁夏", Description: "回族绣花盖头以轻盈的丝绸或纱料制成，边缘绣有精美的花卉和几何图案。盖头的颜色和刺绣风格因地区而异，宁夏地区的盖头以素雅清丽的风格著称，是回族服饰文化的重要符号。",
			ModelFile: "回族绣花盖头.glb",
		},
		{
			Name: "陕西肚兜", Dynasty: "清", Ethnic: "汉族", Craft: "刺绣",
			Material: "丝绸+棉布", Pattern: "五毒纹+莲花纹", CulturalMeaning: "肚兜上绣五毒图案是为了以毒攻毒、驱邪避害，莲花纹象征连生贵子，寄托着长辈对孩童健康成长的深切祝福",
			Province: "陕西", Description: "陕西肚兜是关中地区传统的贴身衣物，通常以大红棉布为底，绣有蝎子、蜈蚣、蟾蜍、壁虎、蛇等五毒图案。色彩鲜艳，针法古朴，是陕西民间刺绣的典型代表。",
			ModelFile: "陕西肚兜.glb",
		},
		{
			Name: "湖北土家族摆手舞衣", Dynasty: "清", Ethnic: "土家族", Craft: "织锦+刺绣",
			Material: "棉布", Pattern: "摆手纹+西兰卡普纹", CulturalMeaning: "摆手舞是土家族最隆重的祭祀舞蹈，舞衣上的图案记录着土家族迁徙和发展的历史，是穿在身上的史诗",
			Province: "湖北", Description: "湖北土家族摆手舞衣是土家族跳摆手舞时穿着的传统服饰，衣身以土家织锦装饰，袖口和领口绣有寓意吉祥的图案。舞衣色彩以红、黑为主，在舞蹈时鲜艳夺目。",
			ModelFile: "湖北土家族摆手舞衣.glb",
		},
		{
			Name: "黎族筒裙", Dynasty: "清", Ethnic: "黎族", Craft: "黎锦织造",
			Material: "棉+丝线", Pattern: "人形纹+蛙纹+大力神纹", CulturalMeaning: "黎锦上的大力神纹是黎族最古老的图腾，人形纹记录着黎族祖先的故事，每一条筒裙都是一部无字的史书",
			Province: "海南", Description: "黎族筒裙以黎锦工艺织造，是中国最古老的纺织技艺之一。黎族妇女使用原始的踞腰织机，以通经断纬技法织出人形纹、动物纹等精美图案，一条筒裙往往需要数月甚至数年才能完成。",
			ModelFile: "黎族筒裙.glb",
		},
		{
			Name: "赫哲族鱼皮衣", Dynasty: "清", Ethnic: "赫哲族", Craft: "鱼皮鞣制+贴花",
			Material: "大马哈鱼皮", Pattern: "云纹+水波纹", CulturalMeaning: "鱼皮衣是人类适应自然的杰出创造，赫哲族以鱼皮为衣的智慧体现了渔猎民族与江河共生共存的深厚情感",
			Province: "黑龙江", Description: "赫哲族鱼皮衣是中国北方渔猎民族赫哲族的独特创造，选用大马哈鱼皮经鞣制、拼接、缝制而成。鱼皮衣轻便防水，表面保留鱼鳞的自然纹理，是一种濒临失传的人类非物质文化遗产。",
			ModelFile: "赫哲族鱼皮衣.glb",
		},
		{
			Name: "朝鲜族短袄裙", Dynasty: "现代", Ethnic: "朝鲜族", Craft: "刺绣",
			Material: "丝绸+棉布", Pattern: "花卉纹+蝴蝶纹", CulturalMeaning: "朝鲜族短袄裙的飘逸线条体现了朝鲜族崇尚自然、追求和谐的美学理念，短袄长裙的搭配展现了女性端庄优雅的气质",
			Province: "吉林", Description: "朝鲜族短袄裙（则高利+契玛）是朝鲜族女性的传统服饰，短袄紧身，长裙高腰，以蝴蝶结系于胸前。色彩明快，线条流畅，是朝鲜族文化的重要标志。",
			ModelFile: "朝鲜族短袄裙.glb",
		},
		{
			Name: "满族旗袍", Dynasty: "清", Ethnic: "满族", Craft: "刺绣+镶滚",
			Material: "丝绸", Pattern: "兰花+蝴蝶纹+海水江崖纹", CulturalMeaning: "旗袍从满族骑射服饰演变而来，承载着满族入关后的文化融合历史，是中国最具国际影响力的传统服饰之一",
			Province: "辽宁", Description: "满族旗袍是旗袍的原始形态，与后来的海派旗袍不同，满族旗袍保留了宽大的直筒形制和精致的镶滚（绲边）工艺。领口、袖口和衣襟处镶有多层滚边，是满族服饰工艺的精华。",
			ModelFile: "满族旗袍.glb",
		},
		{
			Name: "高山族贝珠衣", Dynasty: "清", Ethnic: "高山族", Craft: "贝珠编织",
			Material: "麻布+贝壳珠", Pattern: "几何条纹+人形纹", CulturalMeaning: "贝珠在高山族文化中象征着财富与地位，每一颗贝珠都是手工磨制而成，贝珠衣是高山族献给最尊贵客人的礼物",
			Province: "台湾", Description: "高山族贝珠衣是台湾高山族（原住民）最珍贵的传统服饰，将数万颗手工磨制的贝壳珠缝缀于麻布之上，排列成精美的几何图案。贝珠在阳光下熠熠生辉，是高山族工艺与审美的极致体现。",
			ModelFile: "高山族贝珠衣.glb",
		},
	}

	for _, g := range garments {
		garment := &model.Garment{
			Name:            g.Name,
			Dynasty:         g.Dynasty,
			Ethnic:          g.Ethnic,
			Craft:           g.Craft,
			Material:        g.Material,
			Pattern:         g.Pattern,
			CulturalMeaning: g.CulturalMeaning,
			Province:        g.Province,
			Description:     g.Description,
		}
		id, err := garmentStore.Create(garment)
		if err != nil {
			log.Printf("[Seed] 创建服饰 %s 失败: %v", g.Name, err)
			continue
		}
		log.Printf("[Seed] 服饰 '%s' 创建成功 (ID: %d)", g.Name, id)

		// 尝试关联模型文件
		if g.ModelFile != "" {
			modelPath := filepath.Join("..", "client", "public", "models", g.ModelFile)
			if info, err := os.Stat(modelPath); err == nil {
				modelDir := filepath.Join(cfg.UploadDir, fmt.Sprintf("%d", id))
				os.MkdirAll(modelDir, 0755)

				input, err := os.ReadFile(modelPath)
				if err == nil {
					destPath := filepath.Join(modelDir, g.ModelFile)
					if err := os.WriteFile(destPath, input, 0644); err == nil {
						relPath := filepath.Join(cfg.UploadDir, fmt.Sprintf("%d", id), g.ModelFile)
						m := &model.Model3D{
							Filename:     g.ModelFile,
							OriginalName: g.ModelFile,
							FileSize:     info.Size(),
							MimeType:     "model/gltf-binary",
							FilePath:     relPath,
							GarmentID:    &id,
						}
						mid, err := modelStore.Create(m)
						if err == nil {
							m.ID = mid
							modelStore.Update(mid, m)
							garmentStore.UpdateModelID(id, &mid)
							log.Printf("[Seed]   模型 '%s' 关联成功 (ModelID: %d, Size: %d)", g.ModelFile, mid, info.Size())
						} else {
							log.Printf("[Seed]   模型记录创建失败: %v", err)
						}
					}
				}
			} else {
				log.Printf("[Seed]   模型文件 '%s' 未找到于 %s，跳过关联", g.ModelFile, modelPath)
			}
		}
	}

	// 创建一条每日推荐
	garment, err := garmentStore.GetByID(1)
	if err == nil {
		dailyPickStore.Create(garment.ID, "2026-07-06")
		log.Printf("[Seed] 每日推荐已设置: %s", garment.Name)
	}

	seedAnnotations(db)
	log.Println("[Seed] 种子数据插入完成")
}

// seedAnnotations 为 3 件带模型的示例服饰插入演示热点标注
// 坐标由 tools/seed_hotspots.js 对 GLB 表面射线求交得出（模型居中后、朝向初始镜头 +Z）
func seedAnnotations(db *sql.DB) {
	type demoPoint struct {
		Garment string
		Title   string
		Content string
		X, Y, Z float64
	}
	demos := []demoPoint{
		{Garment: "苗族百褶裙", Title: "百褶工艺", X: 0, Y: 0.2332, Z: 0.1621,
			Content: "百褶裙的裙褶多达数百道，象征苗族祖先迁徙途中翻越的山川，每一道褶皱都是一段迁徙的记忆。"},
		{Garment: "苗族百褶裙", Title: "蜡染底纹", X: 0, Y: -0.0466, Z: 0.2285,
			Content: "以铜刀蘸熔蜡绘于布面，经蓝靛浸染后脱蜡显花，几何纹代表星空与大地，蓝白相间清雅古朴。"},
		{Garment: "苗族百褶裙", Title: "裙摆刺绣", X: 0, Y: -0.3731, Z: 0.2673,
			Content: "花卉纹绣于裙摆之上，寓意生命繁衍，与蜡染底纹相映成趣，是蜡染与刺绣双重工艺的结晶。"},

		{Garment: "苏绣旗袍", Title: "立领盘扣", X: 0, Y: 0.4137, Z: 0.063,
			Content: "传统旗袍立领配手工盘扣，一扣一结皆为心思，勾勒出东方女性的端庄与含蓄。"},
		{Garment: "苏绣旗袍", Title: "花鸟绣面", X: 0, Y: 0.0591, Z: 0.0698,
			Content: "苏绣以「细、雅、洁」著称，花鸟纹样针法细腻、劈丝如发，色彩层次丰富，是中国四大名绣之首的代表。"},
		{Garment: "苏绣旗袍", Title: "旗袍下摆", X: 0, Y: -0.3546, Z: 0.0551,
			Content: "下摆与侧衩处以手工滚边收口，行走间开衩微露，尽显江南水乡的柔美线条。"},

		{Garment: "藏族氆氇袍", Title: "氆氇织造", X: 0, Y: 0.3099, Z: 0.0301,
			Content: "氆氇是藏族最具代表性的手工织品，以羊毛捻线织造，厚实保暖，适应高原多变的气候。"},
		{Garment: "藏族氆氇袍", Title: "几何条纹", X: 0, Y: -0.0516, Z: 0.0563,
			Content: "袍身织入规整的几何条纹，色彩沉稳对比鲜明，体现藏族人民对秩序与自然的审美表达。"},
		{Garment: "藏族氆氇袍", Title: "下摆镶边", X: 0, Y: -0.3615, Z: 0.0519,
			Content: "袍身宽大便于骑马劳作，下摆以同料镶边收束，整体庄重古朴，是高原生活智慧的缩影。"},
	}

	for i, d := range demos {
		var garmentID int64
		err := db.QueryRow(`SELECT id FROM garments WHERE name = ? LIMIT 1`, d.Garment).Scan(&garmentID)
		if err != nil {
			log.Printf("[Seed] 热点标注跳过 '%s': 服饰不存在", d.Garment)
			continue
		}
		_, err = db.Exec(
			`INSERT INTO annotations (garment_id, title, content, position_x, position_y, position_z, sort_order)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			garmentID, d.Title, d.Content, d.X, d.Y, d.Z, i%3,
		)
		if err != nil {
			log.Printf("[Seed] 热点标注 '%s - %s' 插入失败: %v", d.Garment, d.Title, err)
		} else {
			log.Printf("[Seed] 热点标注 '%s - %s' 插入成功", d.Garment, d.Title)
		}
	}
}
