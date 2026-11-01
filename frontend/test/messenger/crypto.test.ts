import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { loadAlphabet, loadDualAlphabets } from '../../src/lib/messenger/crypto/alphabet.ts'
import { fromBase64Url, toBase64Url } from '../../src/lib/messenger/crypto/base64url.ts'
import { compareCodePoints } from '../../src/lib/messenger/crypto/bytes.ts'
import { decryptEnvelope, encryptText, validateDescriptor } from '../../src/lib/messenger/crypto/envelope.ts'
import { generateKeyPair, loadPrivateKey, loadPublicKey } from '../../src/lib/messenger/crypto/keys.ts'
import type { Envelope, PrivateKeyDoc, PublicKeyDoc } from '../../src/lib/messenger/crypto/types.ts'

const read = (path: string) => JSON.parse(readFileSync(new URL(path, import.meta.url), 'utf8'))
const dictionaryOne = read('../../src/lib/messenger/dictionaries/encryption.json')
const dictionaryTwo = read('../../src/lib/messenger/dictionaries/encryption2.json')
const vectors = read('./vectors.json')
const alphabets = await loadDualAlphabets(dictionaryOne, dictionaryTwo)

interface Party {
    private: PrivateKeyDoc
    public: PublicKeyDoc
}

const alice: Party = vectors.alice
const bob: Party = vectors.bob
const partyOf = (direction: string) => (direction === 'alice_to_bob' ? { from: alice, to: bob } : { from: bob, to: alice })

const fixedRandom = (salt: string, nonce: string) => {
    const queue = [fromBase64Url(salt, 'salt'), fromBase64Url(nonce, 'nonce')]

    return () => queue.shift() as Uint8Array
}

test('словари дают те же хэши, что эталонный Python', () => {
    assert.equal(alphabets.first.hash, vectors.alphabet_hash)
    assert.equal(alphabets.second.hash, vectors.alphabet_hash2)
})

for (const vector of vectors.cases) {
    test(`побайтно совпадает с Python: ${vector.name}`, async () => {
        const { from, to } = partyOf(vector.direction)
        const envelope = await encryptText({
            text: vector.text,
            alphabets,
            privateDoc: from.private,
            peerPublicDoc: to.public,
            random: fixedRandom(vector.salt, vector.nonce),
        })

        assert.deepEqual(envelope, vector.envelope)
    })

    test(`расшифровывает конверт Python: ${vector.name}`, async () => {
        const { from, to } = partyOf(vector.direction)
        const text = await decryptEnvelope({ envelope: vector.envelope, alphabets, privateDoc: to.private, peerPublicDoc: from.public })

        assert.equal(text, vector.text)
    })
}

for (const vector of vectors.cases) {
    test(`отправитель перечитывает своё сообщение: ${vector.name}`, async () => {
        const { from, to } = partyOf(vector.direction)
        const text = await decryptEnvelope({ envelope: vector.envelope, alphabets, privateDoc: from.private, peerPublicDoc: to.public, role: 'sender' })

        assert.equal(text, vector.text)
    })
}

test('роль отправителя не даёт читать чужие сообщения', async () => {
    const envelope = await sample()
    const eve = await generateKeyPair()

    await assert.rejects(() => decryptEnvelope({ envelope, alphabets, privateDoc: eve.privateDoc, peerPublicDoc: bob.public, role: 'sender' }), /другим ключом/)
    await assert.rejects(() => decryptEnvelope({ envelope, alphabets, privateDoc: alice.private, peerPublicDoc: eve.publicDoc, role: 'sender' }), /другому получателю/)
})

test('круг шифрование-расшифровка со случайными salt и nonce', async () => {
    const text = 'Привет, мир! 2026-11-01 — проверка.'.replace('—', '-')
    const first = await encryptText({ text, alphabets, privateDoc: alice.private, peerPublicDoc: bob.public })
    const second = await encryptText({ text, alphabets, privateDoc: alice.private, peerPublicDoc: bob.public })

    assert.notEqual(first.ciphertext, second.ciphertext)
    assert.equal(await decryptEnvelope({ envelope: first, alphabets, privateDoc: bob.private, peerPublicDoc: alice.public }), text)
})

test('новая пара ключей работает с векторной стороной', async () => {
    const carol = await generateKeyPair()
    const envelope = await encryptText({ text: 'from carol', alphabets, privateDoc: carol.privateDoc, peerPublicDoc: bob.public })

    assert.equal(envelope.sender_key_id, carol.publicDoc.key_id)
    assert.equal(await decryptEnvelope({ envelope, alphabets, privateDoc: bob.private, peerPublicDoc: carol.publicDoc }), 'from carol')
    assert.equal(carol.publicDoc.public_key.length, 43)
    assert.equal(carol.privateDoc.private_key.length, 43)
})

test('сообщение самому себе', async () => {
    const envelope = await encryptText({ text: 'note to self', alphabets, privateDoc: alice.private, peerPublicDoc: alice.public })

    assert.equal(await decryptEnvelope({ envelope, alphabets, privateDoc: alice.private, peerPublicDoc: alice.public }), 'note to self')
})

const sample = async (): Promise<Envelope> => encryptText({ text: 'secret message', alphabets, privateDoc: alice.private, peerPublicDoc: bob.public })
const open = (envelope: unknown, privateDoc = bob.private, peerPublicDoc = alice.public) => decryptEnvelope({ envelope, alphabets, privateDoc, peerPublicDoc })

test('любая подмена конверта отклоняется', async () => {
    const original = await sample()
    const tokens = original.ciphertext.split(' ')
    const flipped = toBase64Url(fromBase64Url(original.auth_tag, 'tag').map((byte, index) => (index === 0 ? byte ^ 1 : byte)))
    const tampered: Record<string, Envelope> = {
        'токен шифртекста': { ...original, ciphertext: ['AAA', ...tokens.slice(1)].join(' ') },
        'порядок токенов': { ...original, ciphertext: [...tokens].reverse().join(' ') },
        'отрезан хвост': { ...original, ciphertext: tokens.slice(0, -1).join(' ') },
        'auth_tag': { ...original, auth_tag: flipped },
        'salt': { ...original, salt: toBase64Url(new Uint8Array(32)) },
        'nonce': { ...original, nonce: toBase64Url(new Uint8Array(16)) },
    }

    for (const [name, envelope] of Object.entries(tampered)) {
        await assert.rejects(() => open(envelope), `${name} должна ломать расшифровку`)
    }
})

test('чужой ключ и неверный отправитель не расшифровывают', async () => {
    const envelope = await sample()
    const eve = await generateKeyPair()

    await assert.rejects(() => open(envelope, eve.privateDoc), /другого приватного ключа/)
    await assert.rejects(() => open(envelope, bob.private, eve.publicDoc), /неверный публичный ключ/)
})

test('дескриптор: лишние и недостающие поля, версия, алгоритм', async () => {
    const original = await sample()
    const bad = (patch: Record<string, unknown>) => ({ ...original, ...patch })
    const withoutSalt: Record<string, unknown> = { ...original }

    delete withoutSalt.salt

    assert.throws(() => validateDescriptor({ ...original, extra: 1 }), /лишний или недостающий/)
    assert.throws(() => validateDescriptor(withoutSalt), /лишний или недостающий/)
    assert.throws(() => validateDescriptor(bad({ version: 1 })), /версия/)
    assert.throws(() => validateDescriptor(bad({ algorithm: 'AES' })), /Алгоритм/)
    assert.throws(() => validateDescriptor(bad({ type: 'x' })), /type/)
    assert.throws(() => validateDescriptor(bad({ salt: 'a b' })), /salt/)
    assert.throws(() => validateDescriptor(bad({ ciphertext: '' })), /ciphertext/)
    assert.throws(() => validateDescriptor(bad({ sender_key_id: 'short' })), /key_id/)
    assert.throws(() => validateDescriptor(null), /объект/)
    assert.throws(() => validateDescriptor([]), /объект/)
})

test('другой словарь не подходит', async () => {
    const envelope = await sample()
    const swapped = JSON.parse(JSON.stringify(dictionaryOne))

    swapped.US.A = '#!>>'

    const other = await loadDualAlphabets(swapped, dictionaryTwo)

    await assert.rejects(() => decryptEnvelope({ envelope, alphabets: other, privateDoc: bob.private, peerPublicDoc: alice.public }), /другим encryption/)
})

test('символы вне словаря и пустой текст запрещены', async () => {
    for (const text of ['😀', 'a«b', 'tab\t']) {
        await assert.rejects(() => encryptText({ text, alphabets, privateDoc: alice.private, peerPublicDoc: bob.public }), /отсутствует в алфавите/)
    }

    await assert.rejects(() => encryptText({ text: '', alphabets, privateDoc: alice.private, peerPublicDoc: bob.public }), /Пустой текст/)
})

test('токены шифртекста распределены равномерно (шум не выдаёт частоты букв)', async () => {
    const text = 'the quick brown fox jumps over the lazy dog and the cat sat on the mat while the sun was shining '.repeat(30)
    const envelope = await encryptText({ text, alphabets, privateDoc: alice.private, peerPublicDoc: bob.public })
    const tokens = envelope.ciphertext.split(' ')
    const counts = new Map<string, number>()

    for (const token of tokens) {
        counts.set(token, (counts.get(token) ?? 0) + 1)
    }

    const top = Math.max(...counts.values()) / tokens.length
    const uniform = 1 / counts.size

    assert.ok(top < uniform * 2, `самый частый токен ${(top * 100).toFixed(2)}% при равномерных ${(uniform * 100).toFixed(2)}%`)
    assert.ok(tokens.length >= text.length * 4 && tokens.length <= text.length * 7)
})

test('словари: валидация', async () => {
    const clone = () => JSON.parse(JSON.stringify(dictionaryOne))
    const duplicateToken = clone()

    duplicateToken.US.B = duplicateToken.US.A

    const spaceToken = clone()

    spaceToken.US.A = 'a b'

    const emptyToken = clone()

    emptyToken.US.A = ''

    const longKey = clone()

    longKey.US.AB = 'zzz'

    const noSection = clone()

    delete noSection.RU

    const missingChar = clone()

    delete missingChar.US.A

    const duplicateChar = clone()

    duplicateChar.RU.A = 'unique-token-1'

    await assert.rejects(() => loadAlphabet(duplicateToken), /встречается дважды/)
    await assert.rejects(() => loadAlphabet(spaceToken), /пробелы/)
    await assert.rejects(() => loadAlphabet(emptyToken), /Пустой токен/)
    await assert.rejects(() => loadAlphabet(longKey), /одним символом/)
    await assert.rejects(() => loadAlphabet(noSection), /нет секции RU/)
    await assert.rejects(() => loadAlphabet(duplicateChar), /встречается дважды/)
    await assert.rejects(() => loadAlphabet([]), /JSON-объектом/)
    await assert.rejects(() => loadAlphabet(null), /JSON-объектом/)
    await assert.rejects(() => loadDualAlphabets(dictionaryOne, dictionaryOne), /совпадают/)
    await assert.rejects(() => loadDualAlphabets(dictionaryOne, missingChar), /одинаковые символы/)

    const overlapping = JSON.parse(JSON.stringify(dictionaryTwo))

    overlapping.US.A = dictionaryOne.US.A

    await assert.rejects(() => loadDualAlphabets(dictionaryOne, overlapping), /пересекаются/)
})

test('ключи: формат, повреждение и key_id', async () => {
    await assert.rejects(() => loadPrivateKey({ ...alice.private, version: 2 }), /Неподдерживаемый/)
    await assert.rejects(() => loadPrivateKey({ ...alice.private, extra: 1 } as PrivateKeyDoc), /Неподдерживаемый/)
    await assert.rejects(() => loadPrivateKey({ ...alice.private, key_id: 'a'.repeat(32) }), /повреждён/)
    await assert.rejects(() => loadPrivateKey({ ...alice.private, private_key: 'AAAA' }), /32 байта/)
    await assert.rejects(() => loadPublicKey({ ...bob.public, key_id: 'a'.repeat(32) }), /повреждён/)
    await assert.rejects(() => loadPublicKey({ ...bob.public, algorithm: 'RSA' }), /Неподдерживаемый/)
    await assert.rejects(() => loadPublicKey({ ...bob.public, public_key: '!!!' }), /Некорректное поле/)
})

test('base64url и сортировка по кодовым точкам', () => {
    for (const bytes of [[], [0], [255, 254], [1, 2, 3], Array.from({ length: 100 }, (_, index) => index * 2)]) {
        assert.deepEqual(Array.from(fromBase64Url(toBase64Url(new Uint8Array(bytes)), 'x')), bytes)
    }

    for (const bad of ['a', 'abcde', '+++', 'ab=c', 'ab cd']) {
        assert.throws(() => fromBase64Url(bad, 'поле'), /Некорректное поле поле/)
    }

    assert.ok(compareCodePoints('a', 'b') < 0)
    assert.ok(compareCodePoints('b', 'a') > 0)
    assert.equal(compareCodePoints('я', 'я'), 0)
    assert.ok(compareCodePoints('～', '\u{1F600}') < 0)
    assert.ok(compareCodePoints('ab', 'a') > 0)
})

test('резервная копия ключа: круг, неверная фраза, подмена и короткая фраза', async () => {
    const { wrapKey, unwrapKey, WRONG_PASSPHRASE } = await import('../../src/lib/messenger/backup.ts')
    const blob = await wrapKey(alice.private, 'correct horse battery', 1000)
    const parsed = JSON.parse(blob)

    assert.equal(parsed.kdf, 'PBKDF2-SHA256')
    assert.ok(!blob.includes(alice.private.private_key), 'приватный ключ не должен быть виден в копии')
    assert.deepEqual(await unwrapKey(blob, 'correct horse battery'), alice.private)
    assert.notEqual(blob, await wrapKey(alice.private, 'correct horse battery', 1000))

    await assert.rejects(() => unwrapKey(blob, 'wrong passphrase'), new RegExp(WRONG_PASSPHRASE))
    await assert.rejects(() => unwrapKey(JSON.stringify({ ...parsed, ct: parsed.ct.slice(0, -4) + 'AAAA' }), 'correct horse battery'), new RegExp(WRONG_PASSPHRASE))
    await assert.rejects(() => unwrapKey(JSON.stringify({ ...parsed, iter: 999_999_999 }), 'correct horse battery'), new RegExp(WRONG_PASSPHRASE))
    await assert.rejects(() => unwrapKey('not json', 'x'), new RegExp(WRONG_PASSPHRASE))
    await assert.rejects(() => wrapKey(alice.private, 'short', 1000), /короче 8/)
})
