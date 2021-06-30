export const CUSTOM_REGEXES: { [key: string]: CustomRegex } = {
    bic: {
        pattern: '^[A-Z]{3}[A-Z]{3}[A-Z2-9]{1}[A-NP-Z0-9]{1}[A-Z0-9]{3}$',
        validationText: `BIC Code must follow the format of CCCXXXXX000, where:
        • CCC - ISO country code (A-Z): 3 letter code. E.g, for Singapore SGP
        • XXXXXX - first 5 characters of the participant's name. E.g, for MatchMove, MATCH
        • 000 - a unique representation number in WW (0-9): 3-letter code`,
    },

    assetDO: {
        pattern: '^([a-zA-Z]){3}DO$',
        validationText: `Asset Code for a Digital Obligation must end with "DO"`,
    },

    assetDA: {
        pattern: '^([a-zA-Z]){3}$',
        validationText: `Asset Code must be exactly 3 alphabetic characters.`
    }
};

export interface CustomRegex {
    pattern: string;
    validationText: string;
}
