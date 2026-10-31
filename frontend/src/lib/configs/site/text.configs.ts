import { getCurrentLanguage } from "../../i18n/i18n"

const currentLang = getCurrentLanguage()
const langDicts = import.meta.glob<Record<string, string>>("../../meta/*.json", { eager: true })
const langDict = langDicts[`../../meta/${currentLang}.json`];

function TextConfig() {
    return {
        // Header
        title: langDict.title,

        // Header navigation
        NavProduct: langDict.NavProduct,
        NavSecurity: langDict.NavSecurity,
        NavLogin: langDict.NavLogin,
        CreateAccount: langDict.CreateAccount,

        // Home page
        HomeHeroTitle: langDict.HomeHeroTitle,
        HomeHeroText: langDict.HomeHeroText,
        HomeHeroSync: langDict.HomeHeroSync,
        HomeHeroNote: langDict.HomeHeroNote,
        HomeDeviceLaptop: langDict.HomeDeviceLaptop,
        HomeDeviceTablet: langDict.HomeDeviceTablet,
        HomeDevicePhone: langDict.HomeDevicePhone,
        HomeStatusUpdated: langDict.HomeStatusUpdated,
        HomeStatusJustNow: langDict.HomeStatusJustNow,
        HomeTaskEstimate: langDict.HomeTaskEstimate,
        HomeTaskEstimateWhen: langDict.HomeTaskEstimateWhen,
        HomeTaskAccess: langDict.HomeTaskAccess,
        HomeTaskAccessWhen: langDict.HomeTaskAccessWhen,
        HomeTaskReport: langDict.HomeTaskReport,
        HomeTaskActs: langDict.HomeTaskActs,
        HomeTaskActsWhen: langDict.HomeTaskActsWhen,
        HomeTaskEstimateShort: langDict.HomeTaskEstimateShort,
        HomeMoreTasks: langDict.HomeMoreTasks,
        HomeAssigneeAk: langDict.HomeAssigneeAk,
        HomeAssigneeMd: langDict.HomeAssigneeMd,
        HomeAssigneeOp: langDict.HomeAssigneeOp,
        HomeFeatureEncTitle: langDict.HomeFeatureEncTitle,
        HomeFeatureEncText: langDict.HomeFeatureEncText,
        HomeFeatureEuTitle: langDict.HomeFeatureEuTitle,
        HomeFeatureEuText: langDict.HomeFeatureEuText,
        HomeFeatureAuditTitle: langDict.HomeFeatureAuditTitle,
        HomeFeatureAuditText: langDict.HomeFeatureAuditText,

        LoginFailed: langDict.LoginFailed,
        IAgreeToThe: langDict.IAgreeToThe,
        PrivacyPolicy: langDict.PrivacyPolicy,
        TermsOfService: langDict.TermsOfService,
        Password: langDict.Password,
        ExempelPassword: langDict.ExempelPassword,
        DashboardTitle: langDict.DashboardTitle,
        DashboardLogout: langDict.DashboardLogout,
        LanguageTitle: langDict.LanguageTitle,
        LanguageClose: langDict.LanguageClose,
        SidebarCollapse: langDict.SidebarCollapse,
        SidebarExpand: langDict.SidebarExpand,
        SettingsName: langDict.SettingsName,
        SettingsEmail: langDict.SettingsEmail,
        DashboardWelcome: langDict.DashboardWelcome,
        DashboardLoading: langDict.DashboardLoading,
        DashboardLoadFailed: langDict.DashboardLoadFailed,

        // Auth pages (login / register)
        AuthNoAccount: langDict.AuthNoAccount,
        AuthCreateShort: langDict.AuthCreateShort,
        RegisterTitle: langDict.RegisterTitle,
        RegisterIntro: langDict.RegisterIntro,
        RegisterWorkEmail: langDict.RegisterWorkEmail,
        RegisterWorkEmailPlaceholder: langDict.RegisterWorkEmailPlaceholder,
        RegisterTeamName: langDict.RegisterTeamName,
        RegisterTeamNamePlaceholder: langDict.RegisterTeamNamePlaceholder,
        StrengthWeak: langDict.StrengthWeak,
        StrengthFair: langDict.StrengthFair,
        StrengthGood: langDict.StrengthGood,
        StrengthStrong: langDict.StrengthStrong,
        RegisterAnd: langDict.RegisterAnd,
        RegisterCreating: langDict.RegisterCreating,
        RegisterFooterNote: langDict.RegisterFooterNote,
        LoginTitle: langDict.LoginTitle,
        LoginIntro: langDict.LoginIntro,
        LoginEmailLabel: langDict.LoginEmailLabel,
        LoginShowPassword: langDict.LoginShowPassword,
        LoginHidePassword: langDict.LoginHidePassword,
        LoginSubmitting: langDict.LoginSubmitting,
        LoginRemember: langDict.LoginRemember,
        LoginForgot: langDict.LoginForgot,
        LoginSubmit: langDict.LoginSubmit,
        LoginDeviceKey: langDict.LoginDeviceKey,
        OR: langDict.OR,
        LoginPanelTitle: langDict.LoginPanelTitle,
        LoginPanelEncrypted: langDict.LoginPanelEncrypted,
        LoginPanelNote: langDict.LoginPanelNote,
        LoginHash1: langDict.LoginHash1,
        LoginHash2: langDict.LoginHash2,
        LoginHash3: langDict.LoginHash3,
        LoginHash4: langDict.LoginHash4,

        // Menu, dashboard, password
        MenuOpen: langDict.MenuOpen,
        MenuClose: langDict.MenuClose,
        DashboardUserId: langDict.DashboardUserId,
        DashboardCreated: langDict.DashboardCreated,
        PasswordRulesHint: langDict.PasswordRulesHint,
        ProfileTitle: langDict.ProfileTitle,
        ProfileDescription: langDict.ProfileDescription,
        SettingsTitle: langDict.SettingsTitle,
        SettingsDescription: langDict.SettingsDescription,
        SettingsNameHint: langDict.SettingsNameHint,
        SettingsEmailHint: langDict.SettingsEmailHint,
        SettingsSave: langDict.SettingsSave,
        SettingsSaving: langDict.SettingsSaving,
        SettingsSaved: langDict.SettingsSaved,
        SettingsSaveFailed: langDict.SettingsSaveFailed,

        // Register page
        HaveAccount: langDict.HaveAccount,
        RegisterFailed: langDict.RegisterFailed,

        // Errors
        UnexpectedError: langDict.UnexpectedError,
        TraceId: langDict.TraceId,
    }
}

export default TextConfig()