export const languages = ["ru", "en"] as const;
export type Language = (typeof languages)[number];
export function isLanguage(value: string): value is Language {
  return value === "ru" || value === "en";
}

const labels = {
  en: {
    wallets: "Wallets", dashboard: "Dashboard", rates: "Exchange rates", login: "Sign in",
    register: "Create account", logout: "Sign out", email: "Email", password: "Password",
    newWallet: "New wallet", editWallet: "Edit wallet", name: "Wallet name", currency: "Currency",
    save: "Save", deposit: "Deposit", withdraw: "Withdraw", amount: "Amount",
    description: "Description", history: "Recent activity", emptyHistory: "No operations yet.",
    emptyWallets: "Create a wallet to start tracking your balance.", balance: "Balance",
    unavailable: "Couldn't load data. Try again.", requestFailed: "Couldn't save your changes. Check the details and try again.",
    invalidCredentials: "Check your email and password.", emailTaken: "An account already uses this email.",
    insufficient: "There isn't enough money in this wallet.", registerDone: "Account created. Sign in to continue.",
    signedInHelp: "Enter your email and password to open your dashboard.",
    registerHelp: "Create an account to start tracking your money.",
    activity: "Activity", previous: "Previous", next: "Next", heading: "Your wallets",
    summary: "Income and expenses", noReport: "No activity in this period.", addCurrency: "Add currency",
    code: "Currency code", precision: "Fractional digits", lastUpdated: "Last updated",
    base: "Base currency", rate: "Rate", quote: "Quote currency", noRates: "No rates available yet.",
    from: "From", to: "To", filter: "Apply", income: "Income", expenses: "Expenses", allWallets: "All wallets",
    allCurrencies: "All currencies", invalidRange: "Choose a valid date range of up to 366 days.",
  },
  ru: {
    wallets: "Кошельки", dashboard: "Главная", rates: "Курсы валют", login: "Войти",
    register: "Создать аккаунт", logout: "Выйти", email: "Почта", password: "Пароль",
    newWallet: "Новый кошелёк", editWallet: "Редактировать кошелёк", name: "Название кошелька", currency: "Валюта",
    save: "Сохранить", deposit: "Пополнить", withdraw: "Списать", amount: "Сумма",
    description: "Описание", history: "История операций", emptyHistory: "Операций пока нет.",
    emptyWallets: "Создайте кошелёк, чтобы начать отслеживать баланс.", balance: "Баланс",
    unavailable: "Данные недоступны. Попробуйте ещё раз.", requestFailed: "Не удалось сохранить изменения. Проверьте данные и попробуйте ещё раз.",
    invalidCredentials: "Проверьте почту и пароль.", emailTaken: "Аккаунт с этой почтой уже существует.",
    insufficient: "В кошельке недостаточно средств.", registerDone: "Аккаунт создан. Войдите, чтобы продолжить.",
    signedInHelp: "Введите почту и пароль, чтобы открыть главную.",
    registerHelp: "Создайте аккаунт, чтобы отслеживать свои деньги.",
    activity: "Операции", previous: "Назад", next: "Далее", heading: "Ваши кошельки",
    summary: "Доходы и расходы", noReport: "За этот период операций нет.", addCurrency: "Добавить валюту",
    code: "Код валюты", precision: "Знаков после запятой", lastUpdated: "Обновлено",
    base: "Базовая валюта", rate: "Курс", quote: "Валюта котировки", noRates: "Пока нет данных о курсах.",
    from: "С", to: "По", filter: "Применить", income: "Доходы", expenses: "Расходы", allWallets: "Все кошельки",
    allCurrencies: "Все валюты", invalidRange: "Укажите даты за период до 366 дней.",
  },
};
export const dictionary = (language: Language) => labels[language];
