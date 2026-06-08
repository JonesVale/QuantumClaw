#!/usr/bin/env python3
"""
QuantumClaw i18n 全语言翻译脚本
只做增量，不删不改已有翻译
"""
import json, os

BASE = r"H:\AiData\openclaw\workspace\QuantumClaw"
I18N = os.path.join(BASE, "web", "default", "src", "i18n")

# Load reference
with open(os.path.join(I18N, "en.json"), encoding="utf-8") as f:
    en = json.load(f)
with open(os.path.join(I18N, "zh-CN.json"), encoding="utf-8") as f:
    zh_cn = json.load(f)

# ─── Fix zh-CN errors first ───
zh_cn_fixes = {
    "span": "时间跨度",
    "requests_per_window": "每窗口请求数",
    "response_time_ms": "响应时间(毫秒)",
    "MM-DD HH:mm:ss": "MM-DD HH:mm:ss",
    "DeepSeek V3": "DeepSeek V3",
}
zh_cn_path = os.path.join(I18N, "zh-CN.json")
zh_cn_fixed = 0
for k, v in zh_cn_fixes.items():
    if k in zh_cn:
        zh_cn[k] = v
        zh_cn_fixed += 1
zh_cn = dict(sorted(zh_cn.items(), key=lambda x: x[0].lower()))
with open(zh_cn_path, "w", encoding="utf-8", newline="\n") as f:
    json.dump(zh_cn, f, ensure_ascii=False, indent=2)
print(f"zh-CN: fixed {zh_cn_fixed} erroneous translations")

# ─── Common key translations for ALL languages ───
TR = {
    # Identity verification
    "id_number": {
        "zh-TW":"身分證號碼","ja":"ID番号","ko":"신분증 번호","fr":"Numéro d'identité",
        "de":"Ausweisnummer","es":"Número de identidad","it":"Numero ID","ru":"Номер удостоверения",
        "vi":"Số CMND","nl":"ID-nummer","pt":"Número de identidade","ar":"رقم الهوية"
    },
    "id_number_placeholder": {
        "zh-TW":"請輸入您的身分證號碼","ja":"ID番号を入力してください","ko":"신분증 번호를 입력하세요",
        "fr":"Entrez votre numéro d'identité","de":"Geben Sie Ihre Ausweisnummer ein",
        "es":"Ingrese su número de identidad","it":"Inserisci il tuo numero ID",
        "ru":"Введите номер удостоверения","vi":"Nhập số CMND của bạn",
        "nl":"Voer uw ID-nummer in","pt":"Insira seu número de identidade","ar":"أدخل رقم الهوية الخاص بك"
    },
    "identity_verification": {
        "zh-TW":"實名認證","ja":"本人確認","ko":"본인 인증","fr":"Vérification d'identité",
        "de":"Identitätsprüfung","es":"Verificación de identidad","it":"Verifica dell'identità",
        "ru":"Верификация личности","vi":"Xác minh danh tính","nl":"Identiteitsverificatie",
        "pt":"Verificação de identidade","ar":"التحقق من الهوية"
    },
    "identity_verification_desc": {
        "zh-TW":"提現前需要完成實名認證","ja":"出金前に本人確認を完了してください",
        "ko":"출금 전에 본인 인증을 완료해야 합니다","fr":"Vérification d'identité requise avant le retrait",
        "de":"Identitätsprüfung vor Auszahlung erforderlich","es":"Verificación de identidad requerida antes del retiro",
        "it":"Verifica dell'identità richiesta prima del prelievo","ru":"Требуется верификация личности перед выводом",
        "vi":"Xác minh danh tính trước khi rút tiền","nl":"Identiteitsverificatie vereist voor opname",
        "pt":"Verificação de identidade necessária antes do saque","ar":"التحقق من الهوية مطلوب قبل السحب"
    },
    "identity_submitted": {
        "zh-TW":"身份資訊已提交，等待管理員審核","ja":"本人確認情報を送信しました。管理者の確認待ちです",
        "ko":"본인 인증 정보가 제출되었습니다. 관리자 검토 대기 중",
        "fr":"Informations d'identité soumises, en attente de vérification",
        "de":"Identitätsdaten eingereicht, warte auf Prüfung","es":"Información de identidad enviada, pendiente de revisión",
        "it":"Informazioni di identità inviate, in attesa di revisione",
        "ru":"Данные личности отправлены, ожидается проверка","vi":"Thông tin danh tính đã gửi, chờ quản trị viên xem xét",
        "nl":"Identiteitsgegevens ingediend, wacht op beoordeling",
        "pt":"Informações de identidade enviadas, aguardando revisão",
        "ar":"تم إرسال معلومات الهوية، في انتظار مراجعة المسؤول"
    },
    "submit_identity": {
        "zh-TW":"提交審核","ja":"審査に提出","ko":"검토 제출","fr":"Soumettre pour vérification",
        "de":"Zur Prüfung einreichen","es":"Enviar para revisión","it":"Invia per revisione",
        "ru":"Отправить на проверку","vi":"Gửi để xem xét","nl":"Ter beoordeling indienen",
        "pt":"Enviar para revisão","ar":"تقديم للمراجعة"
    },
    "submit_failed": {
        "zh-TW":"提交失敗","ja":"送信に失敗しました","ko":"제출 실패","fr":"Échec de la soumission",
        "de":"Übermittlung fehlgeschlagen","es":"Envío fallido","it":"Invio fallito",
        "ru":"Ошибка отправки","vi":"Gửi thất bại","nl":"Indienen mislukt",
        "pt":"Falha no envio","ar":"فشل الإرسال"
    },
    "real_name": {
        "zh-TW":"姓名","ja":"氏名","ko":"실명","fr":"Nom réel","de":"Echter Name",
        "es":"Nombre real","it":"Nome reale","ru":"Настоящее имя","vi":"Tên thật",
        "nl":"Echte naam","pt":"Nome real","ar":"الاسم الحقيقي"
    },
    "real_name_placeholder": {
        "zh-TW":"請輸入您的真實姓名","ja":"氏名を入力してください","ko":"실명을 입력하세요",
        "fr":"Entrez votre nom réel","de":"Geben Sie Ihren echten Namen ein",
        "es":"Ingrese su nombre real","it":"Inserisci il tuo nome reale",
        "ru":"Введите ваше настоящее имя","vi":"Nhập tên thật của bạn",
        "nl":"Voer uw echte naam in","pt":"Insira seu nome real","ar":"أدخل اسمك الحقيقي"
    },

    # Payment
    "Payment Merchant": {
        "zh-TW":"收款商戶配置","ja":"決済事業者設定","ko":"결제 판매자 설정","fr":"Configuration du commerçant",
        "de":"Händlerkonfiguration","es":"Configuración del comerciante","it":"Configurazione esercente",
        "ru":"Настройки продавца","vi":"Cấu hình người bán","nl":"Handelaarsconfiguratie",
        "pt":"Configuração do comerciante","ar":"إعدادات التاجر"
    },
    "Configure payment providers for user top-up": {
        "zh-TW":"配置支付提供商用於用戶充值","ja":"ユーザーチャージ用の決済プロバイダーを設定",
        "ko":"사용자 충전을 위한 결제 제공업체 구성","fr":"Configurer les fournisseurs de paiement pour le rechargement",
        "de":"Zahlungsanbieter für Benutzeraufladung konfigurieren","es":"Configurar proveedores de pago para recarga",
        "it":"Configura i provider di pagamento per la ricarica","ru":"Настройка платежных провайдеров для пополнения",
        "vi":"Cấu hình nhà cung cấp thanh toán để nạp tiền","nl":"Betalingsproviders configureren voor opwaardering",
        "pt":"Configurar provedores de pagamento para recarga","ar":"تكوين مزودي الدفع لشحن المستخدم"
    },
    "Save Payment Settings": {
        "zh-TW":"保存支付設置","ja":"支払い設定を保存","ko":"결제 설정 저장","fr":"Enregistrer les paramètres de paiement",
        "de":"Zahlungseinstellungen speichern","es":"Guardar configuración de pago","it":"Salva impostazioni di pagamento",
        "ru":"Сохранить настройки оплаты","vi":"Lưu cài đặt thanh toán","nl":"Betalingsinstellingen opslaan",
        "pt":"Salvar configurações de pagamento","ar":"حفظ إعدادات الدفع"
    },
    "Merchant ID": {
        "zh-TW":"商戶ID","ja":"マーチャントID","ko":"판매자 ID","fr":"ID commerçant",
        "de":"Händler-ID","es":"ID de comerciante","it":"ID esercente","ru":"ID продавца",
        "vi":"ID người bán","nl":"Handelaar-ID","pt":"ID do comerciante","ar":"معرف التاجر"
    },
    "Merchant Key": {
        "zh-TW":"商戶密鑰","ja":"マーチャントキー","ko":"판매자 키","fr":"Clé commerçant",
        "de":"Händlerschlüssel","es":"Clave de comerciante","it":"Chiave esercente",
        "ru":"Ключ продавца","vi":"Khóa người bán","nl":"Handelaarssleutel",
        "pt":"Chave do comerciante","ar":"مفتاح التاجر"
    },
    "API Secret": {
        "zh-TW":"API 密鑰","ja":"API シークレット","ko":"API 비밀키","fr":"Secret API",
        "de":"API-Geheimnis","es":"Secreto API","it":"Segreto API","ru":"Секрет API",
        "vi":"Khóa bí mật API","nl":"API-geheim","pt":"Segredo da API","ar":"سر API"
    },
    "Private Key": {
        "zh-TW":"應用私鑰","ja":"アプリケーション秘密鍵","ko":"애플리케이션 개인키",
        "fr":"Clé privée de l'application","de":"Privater Anwendungsschlüssel",
        "es":"Clave privada de la aplicación","it":"Chiave privata dell'applicazione",
        "ru":"Приватный ключ приложения","vi":"Khóa riêng ứng dụng",
        "nl":"Privésleutel van applicatie","pt":"Chave privada do aplicativo","ar":"المفتاح الخاص للتطبيق"
    },
    "Secret Key": {
        "zh-TW":"密鑰","ja":"秘密鍵","ko":"비밀키","fr":"Clé secrète","de":"Geheimschlüssel",
        "es":"Clave secreta","it":"Chiave segreta","ru":"Секретный ключ","vi":"Khóa bí mật",
        "nl":"Geheime sleutel","pt":"Chave secreta","ar":"المفتاح السري"
    },
    "Gateway URL": {
        "zh-TW":"網關地址","ja":"ゲートウェイURL","ko":"게이트웨이 URL","fr":"URL de la passerelle",
        "de":"Gateway-URL","es":"URL de puerta de enlace","it":"URL del gateway",
        "ru":"URL шлюза","vi":"URL cổng","nl":"Gateway-URL","pt":"URL do gateway","ar":"رابط البوابة"
    },
    "Webhook Key": {
        "zh-TW":"Webhook 密鑰","ja":"Webhook キー","ko":"웹훅 키","fr":"Clé Webhook",
        "de":"Webhook-Schlüssel","es":"Clave de Webhook","it":"Chiave Webhook",
        "ru":"Ключ Webhook","vi":"Khóa Webhook","nl":"Webhook-sleutel",
        "pt":"Chave Webhook","ar":"مفتاح Webhook"
    },
    "Sandbox Mode": {
        "zh-TW":"沙箱模式","ja":"サンドボックスモード","ko":"샌드박스 모드","fr":"Mode bac à sable",
        "de":"Sandbox-Modus","es":"Modo sandbox","it":"Modalità sandbox",
        "ru":"Режим песочницы","vi":"Chế độ sandbox","nl":"Sandbox-modus",
        "pt":"Modo sandbox","ar":"وضع الحماية"
    },
    "Min Top-up": {
        "zh-TW":"最小充值","ja":"最低チャージ額","ko":"최소 충전","fr":"Recharge minimum",
        "de":"Mindestaufladung","es":"Recarga mínima","it":"Ricarica minima",
        "ru":"Мин. пополнение","vi":"Nạp tối thiểu","nl":"Minimale opwaardering",
        "pt":"Recarga mínima","ar":"الحد الأدنى للشحن"
    },
    "Transaction Fee": {
        "zh-TW":"交易手續費","ja":"取引手数料","ko":"거래 수수료","fr":"Frais de transaction",
        "de":"Transaktionsgebühr","es":"Comisión de transacción","it":"Commissione di transazione",
        "ru":"Комиссия за транзакцию","vi":"Phí giao dịch","nl":"Transactiekosten",
        "pt":"Taxa de transação","ar":"رسوم المعاملة"
    },
    "Foreign Models": {
        "zh-TW":"國外模型","ja":"海外モデル","ko":"해외 모델","fr":"Modèles étrangers",
        "de":"Ausländische Modelle","es":"Modelos extranjeros","it":"Modelli esteri",
        "ru":"Зарубежные модели","vi":"Mô hình nước ngoài","nl":"Buitenlandse modellen",
        "pt":"Modelos estrangeiros","ar":"النماذج الأجنبية"
    },
    "Domestic Models": {
        "zh-TW":"國內模型","ja":"国内モデル","ko":"국내 모델","fr":"Modèles nationaux",
        "de":"Inländische Modelle","es":"Modelos nacionales","it":"Modelli nazionali",
        "ru":"Отечественные модели","vi":"Mô hình trong nước","nl":"Binnenlandse modellen",
        "pt":"Modelos nacionais","ar":"النماذج المحلية"
    },
    "Foreign Min Fee": {
        "zh-TW":"國外最低手續費","ja":"海外最低手数料","ko":"해외 최소 수수료",
        "fr":"Frais minimum étranger","de":"Mindestgebühr Ausland",
        "es":"Tarifa mínima extranjera","it":"Commissione minima estera",
        "ru":"Мин. комиссия зарубежных","vi":"Phí tối thiểu nước ngoài",
        "nl":"Minimale buitenlandse vergoeding","pt":"Taxa mínima estrangeira",
        "ar":"الحد الأدنى للرسوم الأجنبية"
    },
    "Minimum fee per transaction for foreign models": {
        "zh-TW":"國外模型每筆交易最低手續費","ja":"海外モデルの取引ごとの最低手数料",
        "ko":"해외 모델 거래당 최소 수수료","fr":"Frais minimum par transaction pour les modèles étrangers",
        "de":"Mindestgebühr pro Transaktion für ausländische Modelle",
        "es":"Tarifa mínima por transacción para modelos extranjeros",
        "it":"Commissione minima per transazione per modelli esteri",
        "ru":"Мин. комиссия за транзакцию для зарубежных моделей",
        "vi":"Phí tối thiểu mỗi giao dịch cho mô hình nước ngoài",
        "nl":"Minimale vergoeding per transactie voor buitenlandse modellen",
        "pt":"Taxa mínima por transação para modelos estrangeiros",
        "ar":"الحد الأدنى للرسوم لكل معاملة للنماذج الأجنبية"
    },

    # General UI
    "Official Website": {
        "zh-TW":"官網","ja":"公式サイト","ko":"공식 웹사이트","fr":"Site officiel",
        "de":"Offizielle Website","es":"Sitio web oficial","it":"Sito ufficiale",
        "ru":"Официальный сайт","vi":"Trang web chính thức","nl":"Officiële website",
        "pt":"Site oficial","ar":"الموقع الرسمي"
    },
    "notifications": {
        "zh-TW":"通知","ja":"通知","ko":"알림","fr":"Notifications","de":"Benachrichtigungen",
        "es":"Notificaciones","it":"Notifiche","ru":"Уведомления","vi":"Thông báo",
        "nl":"Meldingen","pt":"Notificações","ar":"الإشعارات"
    },
    "unread": {
        "zh-TW":"未讀","ja":"未読","ko":"읽지 않음","fr":"Non lu","de":"Ungelesen",
        "es":"No leído","it":"Non letto","ru":"Непрочитанные","vi":"Chưa đọc",
        "nl":"Ongelezen","pt":"Não lido","ar":"غير مقروء"
    },
    "used": {
        "zh-TW":"已用","ja":"使用済み","ko":"사용됨","fr":"Utilisé","de":"Verwendet",
        "es":"Usado","it":"Usato","ru":"Использовано","vi":"Đã sử dụng",
        "nl":"Gebruikt","pt":"Usado","ar":"مستخدم"
    },
    "profitable": {
        "zh-TW":"盈利的","ja":"収益性あり","ko":"수익성 있음","fr":"Rentable",
        "de":"Profitabel","es":"Rentable","it":"Redditizio","ru":"Прибыльный",
        "vi":"Có lợi nhuận","nl":"Winstgevend","pt":"Lucrativo","ar":"مربح"
    },
    "negative": {
        "zh-TW":"負面","ja":"マイナス","ko":"부정적","fr":"Négatif","de":"Negativ",
        "es":"Negativo","it":"Negativo","ru":"Отрицательный","vi":"Tiêu cực",
        "nl":"Negatief","pt":"Negativo","ar":"سلبي"
    },
    "settlement_cycle_hint": {
        "zh-TW":"結算週期：每月自動結算上月收入，遇節假日順延","ja":"決済サイクル：毎月自動的に前月の収入を決済、休日の場合は翌営業日",
        "ko":"정산 주기: 매월 자동으로 전월 수익 정산, 공휴일인 경우 다음 영업일로 연기",
        "fr":"Cycle de règlement : règlement automatique mensuel des revenus du mois précédent",
        "de":"Abrechnungszyklus: monatliche automatische Abrechnung der Vormonatseinnahmen",
        "es":"Ciclo de liquidación: liquidación mensual automática de ingresos del mes anterior",
        "it":"Ciclo di regolamento: regolamento automatico mensile delle entrate del mese precedente",
        "ru":"Расчетный цикл: ежемесячный автоматический расчет доходов за предыдущий месяц",
        "vi":"Chu kỳ thanh toán: tự động thanh toán thu nhập tháng trước hàng tháng",
        "nl":"Afrekeningscyclus: maandelijkse automatische afrekening van inkomsten vorige maand",
        "pt":"Ciclo de liquidação: liquidação mensal automática da receita do mês anterior",
        "ar":"دورة التسوية: تسوية تلقائية شهرية لإيرادات الشهر السابق"
    },
    "This is the official reference price. Contact your channel partner for details.": {
        "zh-TW":"此價格為官方指導價，詳情請聯繫您的渠道商。","ja":"こちらは公式参考価格です。詳細はチャネルパートナーにお問い合わせください。",
        "ko":"이 가격은 공식 참고 가격입니다. 자세한 내용은 채널 파트너에게 문의하세요.",
        "fr":"Ceci est le prix de référence officiel. Contactez votre partenaire de canal pour plus de détails.",
        "de":"Dies ist der offizielle Referenzpreis. Kontaktieren Sie Ihren Channel-Partner für Details.",
        "es":"Este es el precio de referencia oficial. Contacte a su socio de canal para más detalles.",
        "it":"Questo è il prezzo di riferimento ufficiale. Contatta il tuo partner di canale per i dettagli.",
        "ru":"Это официальная справочная цена. Свяжитесь с вашим канальным партнером для получения деталей.",
        "vi":"Đây là giá tham khảo chính thức. Liên hệ đối tác kênh của bạn để biết chi tiết.",
        "nl":"Dit is de officiële referentieprijs. Neem contact op met uw kanaalpartner voor details.",
        "pt":"Este é o preço de referência oficial. Entre em contato com seu parceiro de canal para obter detalhes.",
        "ar":"هذا هو السعر المرجعي الرسمي. اتصل بشريك القناة الخاص بك للحصول على التفاصيل."
    },

    # App store badges
    "Apifox Playground": {
        "zh-TW":"Apifox 線上調試","ja":"Apifox プレイグラウンド","ko":"Apifox 플레이그라운드",
        "fr":"Apifox Bac à sable","de":"Apifox Spielwiese","es":"Apifox Área de pruebas",
        "it":"Apifox Area di test","ru":"Apifox Песочница","vi":"Apifox Sân chơi",
        "nl":"Apifox Speeltuin","pt":"Apifox Área de teste","ar":"Apifox الملعب"
    },

    # Quantum
    "ibm-brisbane_recommend": {
        "zh-TW":"適合量子研究和演算法開發","ja":"量子研究とアルゴリズム開発に最適",
        "ko":"양자 연구 및 알고리즘 개발에 적합","fr":"Adapté à la recherche quantique et au développement d'algorithmes",
        "de":"Geeignet für Quantenforschung und Algorithmenentwicklung",
        "es":"Adecuado para investigación cuántica y desarrollo de algoritmos",
        "it":"Adatto per ricerca quantistica e sviluppo di algoritmi",
        "ru":"Подходит для квантовых исследований и разработки алгоритмов",
        "vi":"Phù hợp cho nghiên cứu lượng tử và phát triển thuật toán",
        "nl":"Geschikt voor kwantumonderzoek en algoritmeontwikkeling",
        "pt":"Adequado para pesquisa quântica e desenvolvimento de algoritmos",
        "ar":"مناسب لأبحاث الكم وتطوير الخوارزميات"
    },
    "ibm-brisbane_strengths": {
        "zh-TW":"127量子位元(Eagle處理器)；動態電路；Qiskit運行時","ja":"127量子ビット(Eagleプロセッサ)；動的回路；Qiskitランタイム",
        "ko":"127큐비트(Eagle 프로세서)；동적 회로；Qiskit 런타임",
        "fr":"127 qubits (processeur Eagle)；circuits dynamiques；runtime Qiskit",
        "de":"127 Qubits (Eagle-Prozessor)；dynamische Schaltungen；Qiskit Runtime",
        "es":"127 cúbits (procesador Eagle)；circuitos dinámicos；Qiskit Runtime",
        "it":"127 qubit (processore Eagle)；circuiti dinamici；Qiskit Runtime",
        "ru":"127 кубитов (процессор Eagle)；динамические схемы；Qiskit Runtime",
        "vi":"127 qubit (bộ xử lý Eagle)；mạch động；Qiskit Runtime",
        "nl":"127 qubits (Eagle-processor)；dynamische circuits；Qiskit Runtime",
        "pt":"127 qubits (processador Eagle)；circuitos dinâmicos；Qiskit Runtime",
        "ar":"127 كيوبت (معالج Eagle)；دوائر ديناميكية؛ Qiskit Runtime"
    },
    "ibm-kyiv_recommend": {
        "zh-TW":"專為量子計算研究設計","ja":"量子コンピューティング研究向けに設計",
        "ko":"양자 컴퓨팅 연구를 위해 설계됨","fr":"Conçu pour la recherche en informatique quantique",
        "de":"Entwickelt für Quantencomputing-Forschung","es":"Diseñado para investigación en computación cuántica",
        "it":"Progettato per la ricerca sul calcolo quantistico",
        "ru":"Разработан для исследований в области квантовых вычислений",
        "vi":"Được thiết kế cho nghiên cứu điện toán lượng tử","nl":"Ontworpen voor kwantumcomputeronderzoek",
        "pt":"Projetado para pesquisa em computação quântica","ar":"مصمم لأبحاث الحوسبة الكمومية"
    },
    "ibm-kyiv_strengths": {
        "zh-TW":"127量子位元；IBM量子網路存取","ja":"127量子ビット；IBM Quantum Networkアクセス",
        "ko":"127큐비트；IBM Quantum Network 액세스","fr":"127 qubits；accès IBM Quantum Network",
        "de":"127 Qubits；IBM Quantum Network Zugang","es":"127 cúbits；acceso IBM Quantum Network",
        "it":"127 qubit；accesso IBM Quantum Network","ru":"127 кубитов；доступ к IBM Quantum Network",
        "vi":"127 qubit；truy cập IBM Quantum Network","nl":"127 qubits；IBM Quantum Network toegang",
        "pt":"127 qubits；acesso IBM Quantum Network","ar":"127 كيوبت؛ الوصول إلى شبكة IBM Quantum"
    },
    "ibm-sherbrooke_recommend": {
        "zh-TW":"非常適合錯誤緩解研究","ja":"エラー緩和研究に最適",
        "ko":"오류 완화 연구에 탁월","fr":"Excellent pour la recherche sur l'atténuation des erreurs",
        "de":"Hervorragend für Fehlerminderungsforschung","es":"Excelente para investigación de mitigación de errores",
        "it":"Eccellente per la ricerca sulla mitigazione degli errori",
        "ru":"Отлично подходит для исследований по снижению ошибок",
        "vi":"Tuyệt vời cho nghiên cứu giảm thiểu lỗi","nl":"Uitstekend voor foutmitigatieonderzoek",
        "pt":"Excelente para pesquisa de mitigação de erros","ar":"ممتاز لأبحاث تخفيف الأخطاء"
    },
    "ibm-sherbrooke_strengths": {
        "zh-TW":"127量子位元；錯誤抑制；Q-CTRL整合","ja":"127量子ビット；エラー抑制；Q-CTRL統合",
        "ko":"127큐비트；오류 억제；Q-CTRL 통합","fr":"127 qubits；suppression d'erreurs；intégration Q-CTRL",
        "de":"127 Qubits；Fehlerunterdrückung；Q-CTRL-Integration",
        "es":"127 cúbits；supresión de errores；integración Q-CTRL",
        "it":"127 qubit；soppressione errori；integrazione Q-CTRL",
        "ru":"127 кубитов；подавление ошибок；интеграция Q-CTRL",
        "vi":"127 qubit；triệt tiêu lỗi；tích hợp Q-CTRL",
        "nl":"127 qubits；foutonderdrukking；Q-CTRL-integratie",
        "pt":"127 qubits；supressão de erros；integração Q-CTRL",
        "ar":"127 كيوبت؛ قمع الأخطاء؛ تكامل Q-CTRL"
    },
    "ionq-aria_recommend": {
        "zh-TW":"最適合變分演算法和量子化學","ja":"変分アルゴリズムと量子化学に最適",
        "ko":"변분 알고리즘과 양자 화학에 최적","fr":"Idéal pour les algorithmes variationnels et la chimie quantique",
        "de":"Am besten für Variationsalgorithmen und Quantenchemie","es":"Ideal para algoritmos variacionales y química cuántica",
        "it":"Ideale per algoritmi variazionali e chimica quantistica",
        "ru":"Лучший выбор для вариационных алгоритмов и квантовой химии",
        "vi":"Tốt nhất cho thuật toán biến phân và hóa học lượng tử",
        "nl":"Beste voor variationele algoritmen en kwantumchemie",
        "pt":"Melhor para algoritmos variacionais e química quântica",
        "ar":"الأفضل للخوارزميات التباينية والكيمياء الكمومية"
    },
    "ionq-aria_strengths": {
        "zh-TW":"#1 離子阱量子處理器；25演算法量子位元","ja":"#1 イオントラップ量子プロセッサ；25アルゴリズム量子ビット",
        "ko":"#1 이온 트랩 양자 프로세서；25 알고리즘 큐비트",
        "fr":"Processeur quantique à ions piégés #1；25 qubits algorithmiques",
        "de":"#1 Ionenfallen-Quantenprozessor；25 algorithmische Qubits",
        "es":"Procesador cuántico de iones atrapados #1；25 cúbits algorítmicos",
        "it":"Processore quantistico a ioni intrappolati #1；25 qubit algoritmici",
        "ru":"Квантовый процессор на захваченных ионах #1；25 алгоритмических кубитов",
        "vi":"Bộ xử lý lượng tử ion bẫy #1；25 qubit thuật toán",
        "nl":"#1 ion-val kwantumprocessor；25 algoritmische qubits",
        "pt":"Processador quântico de íons aprisionados #1；25 qubits algorítmicos",
        "ar":"#1 معالج كمي بالأيونات المحاصرة؛ 25 كيوبت خوارزمي"
    },
    "ionq-forte_recommend": {
        "zh-TW":"適合量子機器學習和最佳化","ja":"量子機械学習と最適化に最適",
        "ko":"양자 기계 학습 및 최적화에 이상적","fr":"Idéal pour l'apprentissage automatique quantique et l'optimisation",
        "de":"Ideal für quantenmaschinelles Lernen und Optimierung",
        "es":"Ideal para aprendizaje automático cuántico y optimización",
        "it":"Ideale per machine learning quantistico e ottimizzazione",
        "ru":"Идеально для квантового машинного обучения и оптимизации",
        "vi":"Lý tưởng cho học máy lượng tử và tối ưu hóa",
        "nl":"Ideaal voor kwantummachine learning en optimalisatie",
        "pt":"Ideal para aprendizado de máquina quântico e otimização",
        "ar":"مثالي للتعلم الآلي الكمي والتحسين"
    },
    "ionq-forte_strengths": {
        "zh-TW":"36演算法量子位元；原生閘組；快速電路執行","ja":"36アルゴリズム量子ビット；ネイティブゲートセット；高速回路実行",
        "ko":"36 알고리즘 큐비트；네이티브 게이트 세트；빠른 회로 실행",
        "fr":"36 qubits algorithmiques；jeu de portes natif；exécution rapide de circuits",
        "de":"36 algorithmische Qubits；nativer Gate-Satz；schnelle Schaltungsausführung",
        "es":"36 cúbits algorítmicos；conjunto de puertas nativo；ejecución rápida de circuitos",
        "it":"36 qubit algoritmici；set di porte nativo；esecuzione rapida di circuiti",
        "ru":"36 алгоритмических кубитов；нативный набор вентилей；быстрое выполнение схем",
        "vi":"36 qubit thuật toán；bộ cổng gốc；thực thi mạch nhanh",
        "nl":"36 algoritmische qubits；native gate-set；snelle circuituitvoering",
        "pt":"36 qubits algorítmicos；conjunto de portas nativo；execução rápida de circuitos",
        "ar":"36 كيوبت خوارزمي؛ مجموعة بوابات أصلية؛ تنفيذ سريع للدوائر"
    },
    "ionq-harmony_recommend": {
        "zh-TW":"適合原型開發和教育用途","ja":"プロトタイピングと教育用途に最適",
        "ko":"프로토타이핑 및 교육용으로 완벽","fr":"Parfait pour le prototypage et l'usage éducatif",
        "de":"Perfekt für Prototyping und Bildungszwecke","es":"Perfecto para prototipado y uso educativo",
        "it":"Perfetto per prototipazione e uso didattico","ru":"Идеально для прототипирования и обучения",
        "vi":"Hoàn hảo cho tạo mẫu và sử dụng giáo dục","nl":"Perfect voor prototyping en educatief gebruik",
        "pt":"Perfeito para prototipagem e uso educacional","ar":"مثالي للنماذج الأولية والاستخدام التعليمي"
    },
    "ionq-harmony_strengths": {
        "zh-TW":"11量子位元；穩定運作；教育介面","ja":"11量子ビット；安定動作；教育用インターフェース",
        "ko":"11큐비트；안정적 작동；교육용 인터페이스","fr":"11 qubits；fonctionnement stable；interface éducative",
        "de":"11 Qubits；stabiler Betrieb；Bildungsschnittstelle","es":"11 cúbits；funcionamiento estable；interfaz educativa",
        "it":"11 qubit；funzionamento stabile；interfaccia didattica",
        "ru":"11 кубитов；стабильная работа；образовательный интерфейс",
        "vi":"11 qubit；hoạt động ổn định；giao diện giáo dục",
        "nl":"11 qubits；stabiele werking；educatieve interface",
        "pt":"11 qubits；operação estável；interface educacional",
        "ar":"11 كيوبت؛ تشغيل مستقر؛ واجهة تعليمية"
    },
    "rigetti-ankaa_recommend": {
        "zh-TW":"適合混合經典-量子工作負載","ja":"ハイブリッド古典-量子ワークロードに最適",
        "ko":"하이브리드 클래식-양자 워크로드에 이상적",
        "fr":"Idéal pour les charges de travail hybrides classique-quantique",
        "de":"Ideal für hybride klassisch-quanten Workloads",
        "es":"Ideal para cargas de trabajo híbridas clásico-cuánticas",
        "it":"Ideale per carichi di lavoro ibridi classico-quantistici",
        "ru":"Идеально для гибридных классическо-квантовых нагрузок",
        "vi":"Lý tưởng cho khối lượng công việc lai cổ điển-lượng tử",
        "nl":"Ideaal voor hybride klassiek-kwantum workloads",
        "pt":"Ideal para cargas de trabalho híbridas clássico-quânticas",
        "ar":"مثالي لأحمال العمل الهجينة الكلاسيكية الكمومية"
    },
    "rigetti-ankaa_strengths": {
        "zh-TW":"84量子位元；改進的相干性；增強的閘保真度","ja":"84量子ビット；改善されたコヒーレンス；強化されたゲート忠実度",
        "ko":"84큐비트；개선된 결맞음；향상된 게이트 충실도",
        "fr":"84 qubits；cohérence améliorée；fidélité de porte améliorée",
        "de":"84 Qubits；verbesserte Kohärenz；erhöhte Gate-Treue",
        "es":"84 cúbits；coherencia mejorada；fidelidad de puerta mejorada",
        "it":"84 qubit；coerenza migliorata；fedeltà di gate migliorata",
        "ru":"84 кубита；улучшенная когерентность；повышенная точность вентилей",
        "vi":"84 qubit；độ kết hợp cải thiện；độ trung thực cổng nâng cao",
        "nl":"84 qubits；verbeterde coherentie；verbeterde gate-getrouwheid",
        "pt":"84 qubits；coerência aprimorada；fidelidade de porta aprimorada",
        "ar":"84 كيوبت؛ تماسك محسن؛ دقة بوابة محسنة"
    },
    "rigetti-aspen_recommend": {
        "zh-TW":"適合量子最佳化問題","ja":"量子最適化問題に適しています",
        "ko":"양자 최적화 문제에 적합","fr":"Adapté aux problèmes d'optimisation quantique",
        "de":"Geeignet für Quantenoptimierungsprobleme","es":"Adecuado para problemas de optimización cuántica",
        "it":"Adatto per problemi di ottimizzazione quantistica",
        "ru":"Подходит для задач квантовой оптимизации","vi":"Phù hợp cho các bài toán tối ưu hóa lượng tử",
        "nl":"Geschikt voor kwantumoptimalisatieproblemen",
        "pt":"Adequado para problemas de otimização quântica","ar":"مناسب لمشاكل التحسين الكمومي"
    },
    "rigetti-aspen_strengths": {
        "zh-TW":"80+量子位元；可擴展架構；量子-經典混合","ja":"80+量子ビット；拡張可能なアーキテクチャ；量子-古典ハイブリッド",
        "ko":"80+큐비트；확장 가능한 아키텍처；양자-고전 하이브리드",
        "fr":"80+ qubits；architecture extensible；hybride quantique-classique",
        "de":"80+ Qubits；erweiterbare Architektur；Quanten-Klassik-Hybrid",
        "es":"80+ cúbits；arquitectura extensible；híbrido cuántico-clásico",
        "it":"80+ qubit；architettura estensibile；ibrido quantistico-classico",
        "ru":"80+ кубитов；расширяемая архитектура；квантово-классический гибрид",
        "vi":"80+ qubit；kiến trúc có thể mở rộng；lai lượng tử-cổ điển",
        "nl":"80+ qubits；uitbreidbare architectuur；kwantum-klassiek hybride",
        "pt":"80+ qubits；arquitetura extensível；híbrido quântico-clássico",
        "ar":"80+ كيوبت؛ بنية قابلة للتوسيع؛ هجين كمي كلاسيكي"
    },

    # Brand / Payment names 
    "Epay": {"zh-TW":"易支付","ja":"易支付","ko":"이페이","fr":"Epay","de":"Epay","es":"Epay","it":"Epay","ru":"Epay","vi":"Epay","nl":"Epay","pt":"Epay","ar":"Epay"},
    "Stripe": {"zh-TW":"Stripe（信用卡）","ja":"Stripe（クレジットカード）","ko":"Stripe（신용카드）","fr":"Stripe（carte de crédit）","de":"Stripe（Kreditkarte）","es":"Stripe（tarjeta de crédito）","it":"Stripe（carta di credito）","ru":"Stripe（кредитная карта）","vi":"Stripe（thẻ tín dụng）","nl":"Stripe（creditcard）","pt":"Stripe（cartão de crédito）","ar":"Stripe（بطاقة ائتمان）"},
    "Creem": {"zh-TW":"Creem（聚合支付）","ja":"Creem（統合決済）","ko":"Creem（통합 결제）","fr":"Creem（paiement agrégé）","de":"Creem（aggregierte Zahlung）","es":"Creem（pago agregado）","it":"Creem（pagamento aggregato）","ru":"Creem（агрегированный платеж）","vi":"Creem（thanh toán tổng hợp）","nl":"Creem（geaggregeerde betaling）","pt":"Creem（pagamento agregado）","ar":"Creem（دفع مجمع）"},
    "Waffo": {"zh-TW":"Waffo（跨境支付）","ja":"Waffo（越境決済）","ko":"Waffo（국경 간 결제）","fr":"Waffo（paiement transfrontalier）","de":"Waffo（grenzüberschreitende Zahlung）","es":"Waffo（pago transfronterizo）","it":"Waffo（pagamento transfrontaliero）","ru":"Waffo（трансграничный платеж）","vi":"Waffo（thanh toán xuyên biên giới）","nl":"Waffo（grensoverschrijdende betaling）","pt":"Waffo（pagamento transfronteiriço）","ar":"Waffo（دفع عبر الحدود）"},
    "WorldFirst": {"zh-TW":"萬里匯","ja":"WorldFirst","ko":"WorldFirst","fr":"WorldFirst","de":"WorldFirst","es":"WorldFirst","it":"WorldFirst","ru":"WorldFirst","vi":"WorldFirst","nl":"WorldFirst","pt":"WorldFirst","ar":"WorldFirst"},

    # Misc
    "SLA 99.99%": {"zh-TW":"可用性 99.99%","ja":"可用性 99.99%","ko":"가용성 99.99%","fr":"Disponibilité 99.99%","de":"Verfügbarkeit 99.99%","es":"Disponibilidad 99.99%","it":"Disponibilità 99.99%","ru":"Доступность 99.99%","vi":"Sẵn sàng 99.99%","nl":"Beschikbaarheid 99.99%","pt":"Disponibilidade 99.99%","ar":"توفر 99.99%"},
    "Version 1.0.0": {"zh-TW":"版本 1.0.0","ja":"バージョン 1.0.0","ko":"버전 1.0.0","fr":"Version 1.0.0","de":"Version 1.0.0","es":"Versión 1.0.0","it":"Versione 1.0.0","ru":"Версия 1.0.0","vi":"Phiên bản 1.0.0","nl":"Versie 1.0.0","pt":"Versão 1.0.0","ar":"الإصدار 1.0.0"},
    "app_category_Chat": {"zh-TW":"聊天","ja":"チャット","ko":"채팅","fr":"Chat","de":"Chat","es":"Chat","it":"Chat","ru":"Чат","vi":"Trò chuyện","nl":"Chat","pt":"Chat","ar":"دردشة"},
    "chat_title": {"zh-TW":"聊天","ja":"チャット","ko":"채팅","fr":"Chat","de":"Chat","es":"Chat","it":"Chat","ru":"Чат","vi":"Trò chuyện","nl":"Chat","pt":"Chat","ar":"دردشة"},
    "chat_error_prefix": {"zh-TW":"錯誤","ja":"エラー","ko":"오류","fr":"Erreur","de":"Fehler","es":"Error","it":"Errore","ru":"Ошибка","vi":"Lỗi","nl":"Fout","pt":"Erro","ar":"خطأ"},
    "@/stores/auth-store": {"zh-TW":"@/儲存/身分驗證儲存","ja":"@/ストア/認証ストア","ko":"@/저장소/인증 저장소","fr":"@/stores/auth-store","de":"@/stores/auth-store","es":"@/stores/auth-store","it":"@/stores/auth-store","ru":"@/stores/auth-store","vi":"@/stores/auth-store","nl":"@/stores/auth-store","pt":"@/stores/auth-store","ar":"@/stores/auth-store"},
}

# ─── Apply translations (incremental only!) ───
files = ['zh-TW', 'ja', 'ko', 'fr', 'de', 'es', 'it', 'ru', 'vi', 'nl', 'pt', 'ar']
total_filled = 0

for lang in files:
    fn = f"{lang}.json"
    path = os.path.join(I18N, fn)
    with open(path, encoding="utf-8") as f:
        ld = json.load(f)
    
    filled = 0
    for key, translations in TR.items():
        if lang not in translations:
            continue
        # Only write if: key exists AND current value == en value (untranslated)
        # OR key doesn't exist at all (missing)
        if key in ld:
            if isinstance(ld[key], str) and isinstance(en.get(key), str):
                if ld[key] == en[key] or not ld[key]:
                    ld[key] = translations[lang]
                    filled += 1
        # else: key doesn't exist - skip (don't add new keys beyond the 8 already added)
    
    ld = dict(sorted(ld.items(), key=lambda x: x[0].lower()))
    with open(path, "w", encoding="utf-8", newline="\n") as f:
        json.dump(ld, f, ensure_ascii=False, indent=2)
    
    print(f"[{lang}] filled {filled} translations")
    total_filled += filled

print(f"\nTOTAL: {total_filled} translations applied across 12 languages")
