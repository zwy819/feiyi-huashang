/**
 * seed_annotations.js — 为全部服饰计算三点热点标注的模型表面坐标
 *
 * 复用 seed_hotspots.js 的射线求交思路：读取 GLB 重建世界空间三角形网格，
 * 模型按前端 ModelLoader 的方式平移居中后，从 +Z/-Z 方向（附带少量横向偏移，
 * 兼容鞋类等左右分离的模型）对指定高度比求交，得到表面点。
 *
 * 用法：cd tools && node seed_annotations.js
 * 输出：tools/annotations-plan.json（garment_id → 3 条标注的位置与文案）
 *
 * 入库：配套用 python sqlite3 将 plan 写入 annotations 表
 * （脚本只产出数据，跳过数据库中已有 ≥3 条标注的服饰）。
 */
import { readFileSync, writeFileSync } from 'fs'
import { NodeIO } from '@gltf-transform/core'

const MODELS_DIR = '../client/public/models'

// 三点高度比（相对居中后包围盒：1.0 为顶，0 为底），按服饰类型微调
const H_CLOTH = [0.72, 0.42, 0.08] // 袍服/上衣/连衣裙
const H_SKIRT = [0.62, 0.35, 0.05] // 裙装
const H_FLAT = [0.7, 0.45, 0.12] // 头帕/被面/围裙等平面织物
const H_HAT = [0.6, 0.25, 0.02] // 帽类
const H_SHOE = [0.65, 0.35, 0.05] // 鞋类

/**
 * 全部 40 件服饰。spots 文案与 annotations 表既有 9 条（163/175/187）风格一致；
 * heights 按服饰形制选择三点落位：上段（领口/帽顶/鞋尖）、中段（主体纹样）、下段（下摆/鞋跟/边框）。
 * 已有标注的 163、175、187 也一并输出坐标，入库时由应用端跳过。
 */
const ENTRIES = [
  { id: 163, file: '苗族百褶裙.glb', name: '苗族百褶裙', heights: H_SKIRT },
  { id: 164, file: '苗族银衣.glb', name: '苗族银衣', heights: H_CLOTH, spots: [
    { title: '银角银铃', content: '领口缀满手工打制的银角与银铃，行走时铃音清脆，银饰闪烁如星，是苗族驱邪避凶的护身圣物。' },
    { title: '龙凤银牌', content: '胸前龙凤纹大银牌左右对称，银片间以细链相连，寓意龙凤呈祥，银匠技艺代代相传。' },
    { title: '下摆银坠', content: '下摆银坠成排齐整，整件盛装重达十余公斤，需数年打制，是苗族银饰锻造技艺的集大成之作。' },
  ] },
  { id: 165, file: '侗族亮布衣.glb', name: '侗族亮布衣', heights: H_CLOTH, spots: [
    { title: '亮布光泽', content: '面料经蓝靛反复浸染，再以木槌捶打上万次，布面呈现紫铜色金属光泽，象征光明与希望。' },
    { title: '太阳纹绣片', content: '胸前太阳纹绣片精致醒目，寄托侗族对光明与生命的崇拜，几何纹饰条沿衣襟对称分布。' },
    { title: '几何纹镶边', content: '前襟与下摆以几何纹刺绣镶边收束，与亮布的金属光泽相映，是侗族节庆盛装的点睛之笔。' },
  ] },
  { id: 166, file: '布依族蜡染裙.glb', name: '布依族蜡染裙', heights: H_SKIRT, spots: [
    { title: '蜡染漩涡纹', content: '裙身漩涡纹以铜刀蘸蜡绘就，象征布依族对水的崇拜，蓝靛浸染脱蜡后纹样清雅古朴。' },
    { title: '冰裂纹理', content: '蜡层靛染时自然龟裂，形成独特的冰裂状晕纹，每一条裙子的裂纹都独一无二。' },
    { title: '鱼纹裙摆', content: '裙摆鱼纹饰带寓意丰收与富足，蓝底白花相映成趣，展现布依族与自然的和谐共生。' },
  ] },
  { id: 167, file: '苗族挑花头帕.glb', name: '苗族挑花头帕', heights: H_FLAT, spots: [
    { title: '八角花纹', content: '数纱挑花的八角花纹居中四方连续，八角象征宇宙八方，以布面经纬为坐标全凭心算挑绣。' },
    { title: '蝴蝶纹饰', content: '蝴蝶是苗族创世神话中的万物之母，对称点缀于纹样之间，承载着苗族对祖先的敬仰。' },
    { title: '挑花针脚', content: '针脚呈细密十字形，无需画稿正反两面工整如一，是苗族女性空间观念与智慧的结晶。' },
  ] },
  { id: 168, file: '傣族筒裙.glb', name: '傣族筒裙', heights: H_SKIRT, spots: [
    { title: '织锦孔雀纹', content: '孔雀是傣族的吉祥物，织锦图案绚丽多彩，经纬之间织入傣锦的千年技艺。' },
    { title: '紧身筒形', content: '筒裙紧身合体，行走时裙摆微动，尽显傣家女子的婀娜身姿。' },
    { title: '几何纹饰带', content: '几何纹织带环裙排列，色彩搭配明快，是傣族女性日常与节庆的重要装束。' },
  ] },
  { id: 169, file: '彝族.glb', name: '彝族蜡染百褶裙', heights: H_SKIRT, spots: [
    { title: '蜡染火焰纹', content: '裙身以蜡染绘制火焰纹，象征彝族对火的崇拜与敬畏，蓝白纹样古朴有力。' },
    { title: '贴布绣饰', content: '裙摆以彩色布片贴绣几何纹样，与蜡染底纹相映成趣，双重工艺融为一体。' },
    { title: '百褶裙摆', content: '百褶设计让裙摆展开如孔雀开屏，行走时似火焰摇曳，动感十足。' },
  ] },
  { id: 170, file: '白族扎染上衣.glb', name: '白族扎染上衣', heights: H_CLOTH, spots: [
    { title: '扎染晕纹', content: '纱线捆扎布料后浸入板蓝根染液，解开形成自然晕染，蓝白过渡如云似水。' },
    { title: '蝴蝶花纹', content: '前襟蝴蝶花纹簇拥成趣，每一个扎结都蕴含着白族妇女对生活的美好祝愿。' },
    { title: '几何纹领口', content: '领口与袖口以几何纹边条收束，清新素雅，是大理周城千年扎染技艺的缩影。' },
  ] },
  { id: 171, file: '哈尼族树皮衣.glb', name: '哈尼族树皮衣', heights: H_CLOTH, spots: [
    { title: '树皮纤维', content: '选用构树树皮经浸泡捶打漂洗制成纤维布，质地柔软，是人类最古老的服饰形态之一。' },
    { title: '对襟缝线', content: '对襟无扣以缝线连接裁片，针脚朴拙实用，保留着人类服饰最初的形制。' },
    { title: '原始斑纹', content: '衣面斑驳纹理为树皮天然印记，无染无绣，体现了人与自然最原始的和谐。' },
  ] },
  { id: 172, file: '彝族擦尔瓦.glb', name: '彝族擦尔瓦', heights: H_CLOTH, spots: [
    { title: '羊毛擀毡', content: '优质羊毛擀制成毡，质地厚实保暖，白色象征纯洁与高贵，昼夜为衣为被。' },
    { title: '挑花饰带', content: '胸前几何条纹挑花饰带中央绣日月纹，是彝族人身份与审美的标志。' },
    { title: '百褶毡摆', content: '下摆毡褶细密垂顺，行走时如流水摆动，体现高寒山区的生活智慧。' },
  ] },
  { id: 173, file: '羌族云云鞋.glb', name: '羌族云云鞋', heights: H_SHOE, spots: [
    { title: '云纹鞋面', content: '鞋面满绣彩色云纹，寄托羌族对天的敬仰，针脚细密色彩艳丽。' },
    { title: '上翘鞋尖', content: '鞋尖微微上翘如羊角，羊是羌族图腾，承载着从游牧到农耕的文化记忆。' },
    { title: '挑花鞋帮', content: '鞋帮以精细挑花与刺绣装饰，一针一线皆心意，是羌族服饰文化的重要标志。' },
  ] },
  { id: 174, file: '四川蜀绣嫁衣.glb', name: '四川蜀绣嫁衣', heights: H_CLOTH, spots: [
    { title: '龙凤呈祥', content: '胸前龙凤呈祥绣片以晕针切针绣成，寓意夫妻和谐美满，绣工精细入微。' },
    { title: '芙蓉花纹', content: '芙蓉花是成都的象征，沿前襟与下摆分布，色彩丰富层次分明。' },
    { title: '金线盘绣', content: '金线盘绣镶边勾勒轮廓，与大红绸缎交相辉映，是中国四大名绣的典型代表。' },
  ] },
  { id: 175, file: '藏族氆氇袍.glb', name: '藏族氆氇袍', heights: H_CLOTH },
  { id: 176, file: '藏族邦典围裙.glb', name: '藏族邦典围裙', heights: H_FLAT, spots: [
    { title: '彩虹条纹', content: '横向彩虹条纹细密相间，象征彩虹，是藏族女性成年的标志，每种颜色都是一种祝福。' },
    { title: '氆氇织造', content: '羊毛与丝线交织出细密纹理，日喀则邦典织造技艺已列入国家级非物质文化遗产。' },
    { title: '深色包边', content: '上端配深色腰带便于系束，四缘包边收口，色彩明快而不失稳重。' },
  ] },
  { id: 177, file: '门巴族氆氇袍.glb', name: '门巴族氆氇袍', heights: H_CLOTH, spots: [
    { title: '白色十字纹', content: '白色十字纹象征太阳和四方，体现门巴族原始苯教信仰中对自然的崇拜。' },
    { title: '氆氇织面', content: '羊毛氆氇厚实保暖，与藏式袍服一脉相承又独具门巴特色。' },
    { title: '下摆镶饰', content: '十字纹沿下摆延续收边，袍身宽大，是喜马拉雅南麓高寒生活的智慧结晶。' },
  ] },
  { id: 178, file: '维吾尔艾德莱斯裙.glb', name: '维吾尔艾德莱斯裙', heights: H_SKIRT, spots: [
    { title: '艾德莱斯纹', content: '扎染经线织出如水彩晕染的纹样，巴旦木纹与波浪纹流动其间，无法复制。' },
    { title: '丝绸光泽', content: '和田丝绸经线分段扎染后织造，色彩晕染自然，是丝绸之路上的瑰宝。' },
    { title: '裙摆飘逸', content: '裙摆宽大飘逸，行走时色彩流动如沙漠绿洲，展现维吾尔族的热情与创造力。' },
  ] },
  { id: 179, file: '哈萨克族绣花大衣.glb', name: '哈萨克族绣花大衣', heights: H_CLOTH, spots: [
    { title: '羊角纹刺绣', content: '领口袖口羊角纹刺绣精美，象征牛羊成群、生活富足。' },
    { title: '贴花工艺', content: '厚实呢料之上贴花与刺绣并用，色彩明艳对比强烈，尽显草原豪迈。' },
    { title: '鹰纹下摆', content: '下摆纹样融入鹰图腾元素，鹰是哈萨克族精神的象征，展翅欲飞。' },
  ] },
  { id: 180, file: '柯尔克孜族白毡帽.glb', name: '柯尔克孜族白毡帽', heights: H_HAT, spots: [
    { title: '卷草帽顶', content: '帽顶绣卷草纹圆花，白色象征雪山般纯洁的品格，毡面细腻洁白。' },
    { title: '几何纹帽身', content: '帽身一圈几何纹刺绣精美，是柯尔克孜族民族认同的重要标志。' },
    { title: '上翻帽檐', content: '帽檐两侧微微上翻如天山雪峰，御寒保暖之余更显巍峨气度。' },
  ] },
  { id: 181, file: '壮族织锦被面.glb', name: '壮族织锦被面', heights: H_FLAT, spots: [
    { title: '壮锦图案', content: '通经断纬织出菱形纹、蛙纹与太阳纹，与云锦、蜀锦、宋锦并称中国四大名锦。' },
    { title: '蛙纹图腾', content: '蛙是壮族的图腾，象征雨水与丰收，太阳纹代表光明与生命。' },
    { title: '织造结构', content: '棉线为经、彩色丝线为纬，结构严谨色彩对比强烈，织造技艺精湛。' },
  ] },
  { id: 182, file: '瑶族刺绣盘王衣.glb', name: '瑶族刺绣盘王衣', heights: H_CLOTH, spots: [
    { title: '盘王印图案', content: '通身绣满盘王印，是瑶族子孙认祖归宗的标志，承载迁徙的历史记忆。' },
    { title: '龙犬纹饰', content: '龙犬纹讲述盘瓠传说，绣工精细繁复，一件盛装往往需数年才能完成。' },
    { title: '祭祀盛装', content: '瑶族祭祀盘王时的礼服，通身刺绣繁复华丽，是瑶族刺绣技艺的最高体现。' },
  ] },
  { id: 183, file: '土家族织锦衣裙.glb', name: '土家族织锦衣裙', heights: H_CLOTH, spots: [
    { title: '台台花纹', content: '台台花是土家族对虎的抽象表达，体现了白虎图腾崇拜的悠久传统。' },
    { title: '阳雀花纹', content: '阳雀花象征报春与希望，以古老斜织腰机通经断纬织就。' },
    { title: '西兰卡普', content: '土家织锦图案多达百余种，色彩古朴厚重，被誉为纺织业的活化石。' },
  ] },
  { id: 184, file: '苗族花带裙.glb', name: '苗族花带裙', heights: H_SKIRT, spots: [
    { title: '花带拼接', content: '数十条手工花带拼接成裙，每条宽仅数厘米却织满精美繁复的几何图案。' },
    { title: '八角花纹', content: '花带上的八角花与菱形纹交织，蕴含着编织者的祝福和心意。' },
    { title: '裙摆飘带', content: '行走时花带飘动色彩斑斓，花带是湘西苗族女性传递情感的媒介。' },
  ] },
  { id: 185, file: '蒙古族长袍.glb', name: '蒙古族长袍', heights: H_CLOTH, spots: [
    { title: '镶边工艺', content: '领口袖口镶边精美，适应草原游牧生活的需要，兼具实用性与艺术性。' },
    { title: '云纹刺绣', content: '云纹与动物纹绣于袍身，展现蒙古族对草原和自然的敬畏。' },
    { title: '宽袍大袖', content: '宽大版型便于骑射劳作，尽显游牧民族的豪迈气概。' },
  ] },
  { id: 186, file: '蒙古族喀尔喀袍.glb', name: '蒙古族喀尔喀袍', heights: H_CLOTH, spots: [
    { title: '如意纹镶边', content: '多层镶边绣如意纹与盘肠纹，每一道镶边都代表着一重吉祥祝福。' },
    { title: '贴花刺绣', content: '贴花与刺绣并用，盘肠纹寓意吉祥绵延不绝，工艺考究。' },
    { title: '隆起肩部', content: '肩部隆起如山峰，象征蒙古族对圣山的崇拜，整体端庄大气。' },
  ] },
  { id: 187, file: '苏绣旗袍.glb', name: '苏绣旗袍', heights: H_CLOTH },
  { id: 188, file: '汉族缂丝吉服.glb', name: '汉族缂丝吉服', heights: H_CLOTH, spots: [
    { title: '缂丝工艺', content: '通经断纬，图案与面料一体成型，素有「一寸缂丝一寸金」之说。' },
    { title: '云龙纹样', content: '花卉纹与云龙纹织入吉服，承载着对穿着者的最高祝福与尊崇。' },
    { title: '金线织造', content: '金线点缀其间，一件吉服往往需数年织成，是皇家贵族的专属奢华。' },
  ] },
  { id: 189, file: '汉族杭绣礼服.glb', name: '汉族杭绣礼服', heights: H_CLOTH, spots: [
    { title: '西湖景致', content: '以西湖山水为灵感源泉，将江南的诗情画意融入一针一线。' },
    { title: '花卉纹绣', content: '针法细腻多变，色彩清雅秀丽，传承南宋宫廷刺绣遗风。' },
    { title: '礼服形制', content: '剪裁得体端庄大方，与苏绣、湘绣齐名，是江南刺绣的重要流派。' },
  ] },
  { id: 190, file: '畲族凤凰装.glb', name: '畲族凤凰装', heights: H_CLOTH, spots: [
    { title: '凤凰纹绣', content: '凤凰是畲族图腾，传说畲族始祖是凤凰化身，穿上凤凰装即获凤凰的祝福。' },
    { title: '牡丹纹饰', content: '牡丹纹与凤凰纹相映成趣，红色主调艳丽喜庆，是畲族女性的婚礼盛装。' },
    { title: '腰间织带', content: '腰系手工织带，与凤凰冠配套穿着，展现畲族独特的审美追求。' },
  ] },
  { id: 191, file: '广东潮绣婚服.glb', name: '广东潮绣婚服', heights: H_CLOTH, spots: [
    { title: '垫高绣法', content: '潮绣特有的垫高绣法使绣面立体如浮雕，金线银线交叠，金碧辉煌。' },
    { title: '龙凤盘金', content: '金线银线盘绣龙凤与牡丹，象征婚姻的珍贵与永恒。' },
    { title: '婚服形制', content: '大红婚服富丽堂皇，是粤绣重要分支潮绣的巅峰之作。' },
  ] },
  { id: 192, file: '福建惠安女服饰.glb', name: '福建惠安女服饰', heights: H_CLOTH, spots: [
    { title: '花头巾', content: '黄斗笠配花头巾，遮阳挡风，是惠安女最具辨识度的经典装扮。' },
    { title: '短上衣', content: '紧身短上衣配色大胆鲜艳，「节约衣」之名由此而来，造型独特。' },
    { title: '宽腿裤', content: '裤腿异常宽大，便于滩涂劳作，体现闽南女性勤劳坚韧的品格。' },
  ] },
  { id: 193, file: '藏族安多袍.glb', name: '藏族安多袍', heights: H_CLOTH, spots: [
    { title: '羔皮镶边', content: '领口、袖口和下摆镶有名贵的羔皮，是安多藏族身份的象征。' },
    { title: '氆氇袍身', content: '厚实氆氇面料适应高原多变气候，束腰后形成自然的兜囊。' },
    { title: '宽袍形制', content: '袍身宽大便于游牧携带物品，是高原生活智慧的体现。' },
  ] },
  { id: 194, file: '裕固族红缨帽.glb', name: '裕固族红缨帽', heights: H_HAT, spots: [
    { title: '红色流苏', content: '帽顶红色流苏象征太阳，是裕固族女性最具特色的头饰。' },
    { title: '珊瑚珠绣', content: '帽身绣满彩色丝线与珊瑚珠，珊瑚珠代表珍贵与吉祥。' },
    { title: '尖顶帽身', content: '尖顶帽身以白羊毛擀制，造型反映裕固族从西域迁徙而来的渊源。' },
  ] },
  { id: 195, file: '回族绣花盖头.glb', name: '回族绣花盖头', heights: H_FLAT, spots: [
    { title: '刺绣花纹', content: '边缘绣有花卉与几何纹样，融合中国传统纹样与伊斯兰文化元素。' },
    { title: '轻盈纱料', content: '丝绸纱料轻盈飘逸，宁夏盖头以素雅清丽的风格著称。' },
    { title: '盖头形制', content: '盖头是回族女性信仰与美的结合，是回族服饰文化的重要符号。' },
  ] },
  { id: 196, file: '陕西肚兜.glb', name: '陕西肚兜', heights: H_CLOTH, spots: [
    { title: '五毒纹样', content: '绣蝎子、蜈蚣、蟾蜍、壁虎、蛇五种毒虫，寓意以毒攻毒、驱邪避害。' },
    { title: '莲花纹饰', content: '莲花纹象征连生贵子，寄托着长辈对孩童健康成长的深切祝福。' },
    { title: '贴身形制', content: '大红棉布为底贴身穿着，针法古朴，是关中民间刺绣的典型代表。' },
  ] },
  { id: 197, file: '湖北土家族摆手舞衣.glb', name: '湖北土家族摆手舞衣', heights: H_CLOTH, spots: [
    { title: '摆手纹饰', content: '舞衣图案记录着土家族迁徙和发展的历史，是穿在身上的史诗。' },
    { title: '西兰卡普纹', content: '衣身以土家织锦装饰，红黑主调在摆手舞中鲜艳夺目。' },
    { title: '袖口领口绣', content: '袖口和领口绣有寓意吉祥的图案，供摆手舞祭祀庆典穿着。' },
  ] },
  { id: 198, file: '黎族筒裙.glb', name: '黎族筒裙', heights: H_SKIRT, spots: [
    { title: '大力神纹', content: '大力神纹是黎族最古老的图腾，以通经断纬技法织入裙身。' },
    { title: '人形纹饰', content: '人形纹记录着黎族祖先的故事，一条筒裙就是一部无字的史书。' },
    { title: '踞腰织机', content: '以原始踞腰织机织造，一条筒裙往往需要数月甚至数年才能完成。' },
  ] },
  { id: 199, file: '赫哲族鱼皮衣.glb', name: '赫哲族鱼皮衣', heights: H_CLOTH, spots: [
    { title: '鱼鳞纹理', content: '大马哈鱼皮保留天然鱼鳞纹理，轻便防水，独具渔猎民族匠心。' },
    { title: '鞣制工艺', content: '鱼皮经鞣制、拼接、缝制而成，是濒临失传的人类非物质文化遗产。' },
    { title: '云纹贴花', content: '云纹与水波纹贴花装饰，体现赫哲族与江河共生共存的深厚情感。' },
  ] },
  { id: 200, file: '朝鲜族短袄裙.glb', name: '朝鲜族短袄裙', heights: H_CLOTH, spots: [
    { title: '蝴蝶结带', content: '胸前蝴蝶结系带是朝鲜族服饰的经典细节，温婉秀美。' },
    { title: '短袄形制', content: '短袄「则高利」紧身合体，线条流畅，尽显端庄优雅气质。' },
    { title: '高腰长裙', content: '长裙「契玛」高腰飘逸，色彩明快，体现朝鲜族崇尚自然的美学。' },
  ] },
  { id: 201, file: '满族旗袍.glb', name: '满族旗袍', heights: H_CLOTH, spots: [
    { title: '多层滚边', content: '领口、袖口和衣襟镶有多层滚边，一针一线皆是满族服饰工艺的精华。' },
    { title: '海水江崖纹', content: '下摆海水江崖纹寓意山河永固，承载满族骑射服饰演变的历史。' },
    { title: '直筒形制', content: '保留宽大直筒的原始形态，与后来的海派旗袍风格迥异。' },
  ] },
  { id: 202, file: '高山族贝珠衣.glb', name: '高山族贝珠衣', heights: H_CLOTH, spots: [
    { title: '贝珠缀织', content: '数万颗手工磨制的贝壳珠缝缀于麻布之上，阳光下熠熠生辉。' },
    { title: '几何条纹', content: '贝珠排列成精美的几何条纹与人形纹，象征财富与地位。' },
    { title: '礼服形制', content: '贝珠衣是高山族献给最尊贵客人的礼物，是工艺与审美的极致体现。' },
  ] },
]

// ---------- 精简 vec3 / mat4 工具（与 seed_hotspots.js 一致） ----------
const v3 = (x = 0, y = 0, z = 0) => [x, y, z]
const sub = (a, b) => [a[0] - b[0], a[1] - b[1], a[2] - b[2]]
const cross = (a, b) => [
  a[1] * b[2] - a[2] * b[1],
  a[2] * b[0] - a[0] * b[2],
  a[0] * b[1] - a[1] * b[0],
]
const dot = (a, b) => a[0] * b[0] + a[1] * b[1] + a[2] * b[2]

function composeTRS(t, q, s) {
  const [x, y, z, w] = q
  const x2 = x + x, y2 = y + y, z2 = z + z
  const xx = x * x2, xy = x * y2, xz = x * z2
  const yy = y * y2, yz = y * z2, zz = z * z2
  const wx = w * x2, wy = w * y2, wz = w * z2
  return [
    (1 - (yy + zz)) * s[0], (xy + wz) * s[0], (xz - wy) * s[0], 0,
    (xy - wz) * s[1], (1 - (xx + zz)) * s[1], (yz + wx) * s[1], 0,
    (xz + wy) * s[2], (yz - wx) * s[2], (1 - (xx + yy)) * s[2], 0,
    t[0], t[1], t[2], 1,
  ]
}
function mul4(a, b) {
  const out = new Array(16).fill(0)
  for (let c = 0; c < 4; c++)
    for (let r = 0; r < 4; r++)
      for (let k = 0; k < 4; k++) out[c * 4 + r] += a[k * 4 + r] * b[c * 4 + k]
  return out
}
function transformPoint(m, p) {
  return [
    m[0] * p[0] + m[4] * p[1] + m[8] * p[2] + m[12],
    m[1] * p[0] + m[5] * p[1] + m[9] * p[2] + m[13],
    m[2] * p[0] + m[6] * p[1] + m[10] * p[2] + m[14],
  ]
}
const I4 = [1,0,0,0, 0,1,0,0, 0,0,1,0, 0,0,0,1]

function rayTriangle(orig, dir, tri) {
  const [v0, v1, v2] = tri
  const e1 = sub(v1, v0)
  const e2 = sub(v2, v0)
  const p = cross(dir, e2)
  const det = dot(e1, p)
  if (Math.abs(det) < 1e-9) return -1
  const invDet = 1 / det
  const tvec = sub(orig, v0)
  const u = dot(tvec, p) * invDet
  if (u < 0 || u > 1) return -1
  const q = cross(tvec, e1)
  const v = dot(dir, q) * invDet
  if (v < 0 || u + v > 1) return -1
  const t = dot(e2, q) * invDet
  return t > 1e-6 ? t : -1
}

async function loadTriangles(path) {
  const io = new NodeIO()
  const doc = await io.read(path)
  const tris = []
  for (const scene of doc.getRoot().listScenes()) {
    const walk = (node, parentMat) => {
      const world = mul4(parentMat, composeTRS(node.getTranslation(), node.getRotation(), node.getScale()))
      const mesh = node.getMesh()
      if (mesh) {
        for (const prim of mesh.listPrimitives()) {
          const pos = prim.getAttribute('POSITION')
          if (!pos) continue
          const idx = prim.getIndices()
          const count = idx ? idx.getCount() : pos.getCount()
          const el = [0, 0, 0]
          const pt = (i) => transformPoint(world, pos.getElement(i, el))
          for (let i = 0; i + 2 < count; i += 3) {
            const ia = idx ? idx.getScalar(i) : i
            const ib = idx ? idx.getScalar(i + 1) : i + 1
            const ic = idx ? idx.getScalar(i + 2) : i + 2
            tris.push([pt(ia), pt(ib), pt(ic)])
          }
        }
      }
      for (const child of node.listChildren()) walk(child, world)
    }
    for (const node of scene.listChildren()) walk(node, I4)
  }
  return tris
}

function computeBBox(tris) {
  const min = v3(Infinity, Infinity, Infinity)
  const max = v3(-Infinity, -Infinity, -Infinity)
  for (const tri of tris)
    for (const v of tri)
      for (let i = 0; i < 3; i++) {
        min[i] = Math.min(min[i], v[i])
        max[i] = Math.max(max[i], v[i])
      }
  return { min, max }
}

function centerTriangles(tris, center) {
  for (const tri of tris) for (const v of tri) {
    v[0] -= center[0]; v[1] -= center[1]; v[2] -= center[2]
  }
}

/**
 * 在指定高度取表面点：从 +Z 优先（初始相机方向），失败后尝试 -Z，
 * 并附带少量横向偏移，兼容鞋类等左右分离的模型；
 * 仍不中时退化为"目标高度附近 z 最大的顶点"（应对银坠间隙、窄小鞋体等）。
 */
function surfacePointAtY(tris, y, maxAbsX) {
  const radius = 50
  const offsets = maxAbsX > 0 ? [0, maxAbsX * 0.25, -maxAbsX * 0.25] : [0]
  for (const dirZ of [1, -1]) {
    for (const offX of offsets) {
      const orig = v3(offX, y, dirZ * radius)
      const dir = v3(0, 0, -dirZ)
      let bestT = Infinity
      for (const tri of tris) {
        const t = rayTriangle(orig, dir, tri)
        if (t > 0 && t < bestT) bestT = t
      }
      if (bestT < Infinity) {
        return [
          orig[0] + dir[0] * bestT,
          orig[1] + dir[1] * bestT,
          orig[2] + dir[2] * bestT,
        ]
      }
    }
  }
  // 顶点兜底：取 |vy - y| 最小（容差内取 z 最大）的真实顶点
  let best = null
  let bestDy = Infinity
  for (const tri of tris) {
    for (const v of tri) {
      const dy = Math.abs(v[1] - y)
      if (dy < bestDy - 1e-6 || (dy < bestDy + 1e-6 && best && v[2] > best[2])) {
        bestDy = dy
        best = v
      }
    }
  }
  return best ? [best[0], best[1], best[2]] : null
}

const plan = {}
const failures = []
for (const entry of ENTRIES) {
  const path = `${MODELS_DIR}/${entry.file}`
  let tris
  try {
    tris = await loadTriangles(path)
  } catch (err) {
    failures.push(`${entry.name}: 读取失败 ${err.message}`)
    continue
  }
  if (!tris.length) {
    failures.push(`${entry.name}: 无三角形`)
    continue
  }
  const { min, max } = computeBBox(tris)
  const center = [(min[0] + max[0]) / 2, (min[1] + max[1]) / 2, (min[2] + max[2]) / 2]
  centerTriangles(tris, center)
  const size = sub(max, min)
  const maxAbsX = Math.max(Math.abs(min[0] - center[0]), Math.abs(max[0] - center[0]))

  // 未提供文案的既有三件（163/175/187）只输出坐标，入库时跳过
  const texts = entry.spots
  const spots = []
  for (let i = 0; i < entry.heights.length; i++) {
    const ratio = entry.heights[i]
    const y = -size[1] / 2 + size[1] * ratio
    const p = surfacePointAtY(tris, y, maxAbsX)
    spots.push({
      sort_order: i,
      title: texts ? texts[i].title : null,
      content: texts ? texts[i].content : null,
      position: p ? p.map((n) => Number(n.toFixed(4))) : null,
    })
    if (!p) failures.push(`${entry.name}: 高度 ${ratio} 未命中表面`)
  }
  plan[entry.id] = { name: entry.name, size: size.map((n) => Number(n.toFixed(3))), spots }
}

writeFileSync('annotations-plan.json', JSON.stringify(plan, null, 2), 'utf8')
const total = Object.values(plan).reduce((n, e) => n + e.spots.length, 0)
console.log(`完成：${Object.keys(plan).length} 件服饰，共 ${total} 个标注点 → annotations-plan.json`)
if (failures.length) {
  console.error('警告：')
  for (const f of failures) console.error('  ' + f)
}
