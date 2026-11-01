export const KEY_VERSION = 1
export const ENVELOPE_VERSION = 2
export const ENVELOPE_TYPE = 'my-masengers/message'
export const ALGORITHM = 'X25519-HKDF-SHA256+DualAlphabetNoise-FullRandom-v2'
export const MAX_TEXT_BYTES = 1_048_576

export interface PublicKeyDoc {
    version: number
    algorithm: string
    key_id: string
    public_key: string
}

export interface PrivateKeyDoc extends PublicKeyDoc {
    private_key: string
}

export interface Envelope {
    type: string
    version: number
    algorithm: string
    alphabet_hash: string
    alphabet_hash2: string
    sender_key_id: string
    recipient_key_id: string
    salt: string
    nonce: string
    ciphertext: string
    auth_tag: string
}

export interface Alphabet {
    forward: Map<string, string>
    hash: string
}

export type RandomBytes = (length: number) => Uint8Array