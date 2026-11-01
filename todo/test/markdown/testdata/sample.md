# PureC-OS todo

## UserSpace

- [x] Доделать нормально мишку | priority:normal
- [x] Добавить поддержку Usb Mouse в userspace | priority:normal
- [x] Исправить Баги с мышкой | priority:normal
- [ ] А так же будет сделана миграция UserSpace в Ring 3. | priority:normal
- [x] User space ведёт себя не так как задумано | priority:high | tags:bug | deadline:2026-09-23
  Когда userspace а точнее window manager стартует в отдельном процессе Я открываю окно что-то делаю в программе потом закрываю окно
  И полностью вся система падает с какой ошибкой я не знаю
  Просто зависает

## Userspace programs

- [x] Портировать Lua 5.4 REPL (v1, без math): вендор, hosted libc, сборка, манифест | priority:normal
  - [x] Эхо ввода в REPL (purec_lua_readline): посимвольное эхо + backspace, проверено живьём в терминале | priority:normal

## Network

- [x] Добавить поддержку 802.11 ассоциации | priority:high
  Частично было добавлено
  - [x] Добавить поддержку 802.11b | priority:normal
- [ ] Сделать норм wifi backend | priority:normal
  - [ ] Добавить нормальный бэкенд для сети | priority:normal
- [x] Поднять Atheros AR928X Wirless Network Adapter (Этот пункт частично реализован, но не работаеть)| priority:normal | priority:normal
- [ ] нужно снимать rfkill с ar928 | priority:low | tags:dev
  нужно доделать ar928 начать снимать rfkill и если там есть пропреитарние блобы их нужно грузить

## Kernel

- [x] Добавить менеджер устройств через ACPI чтения AML namespace | priority:normal
  - [x] чтение всех устройств и их свойств | priority:normal
  - [ ] Добавить возможность управления устройствами | priority:normal
- [x] Сделать возможность чтения того что можно прочитать из ACPI | priority:normal
  - [ ] Сделать возможность управления вентиляторами через ACPI | priority:normal
    - [ ] Скорее всего будет через vender soft | priority:normal

## Touchpad

- [ ] Провести тестирование драйвера touchpad на BareMetal | priority:high
  Частично было проверено и тачпад не работает

## SMP и Много процесорность

- [x] Синхронизация | priority:normal
  - [x] Защита общих структур: | priority:normal
    scheduler — IRQ-safe spinlock, передача блокировки через переключение стека process table — mutex с IRQ-safe защитой метаданных VMM / page tables — IRQ-safe spinlock и подтверждаемый TLB shootdown PMM — IRQ-safe spinlock VFS — рекурсивный mutex, допускающий сон при I/O
  - [x] Добавить cli/sti аккуратно + memory barriers где нужно | priority:normal

## SO формат

- [/] Ревизия стандартов линковки ядра (linker-limine.ld, toolchain.mk, MOD_CFLAGS) | priority:high
- [ ] Спецификация формата .so для PureC-OS (ET_DYN, PT_DYNAMIC, baselib) | priority:high
