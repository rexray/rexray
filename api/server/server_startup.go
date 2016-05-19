package server

const serverStartupHeader = `
                       ''''.--.'''
                 ''''.-:/+oyyyso+:-.'''
            '''..:/+osyyyyyyyyyyyyyso+/-..'''
      '''..-/+osyyyyyyyyyyyyyyyyyyyyyyyyso+/-.'''
  ''.-:/osyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyso+/-..''
'.-://osyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyso+:-.''
..++++/////+osyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyysso++ooo..
.:oooo++++++////+osyyyyyyyyyyyyyyyyyyyyyyyssooooosyhhddh.'
./ooooooooooo++++////+osyyyyyyyyyyyyssoo+ooshhddhddddddh.'
.-oosssssssssoooooo+++////+ossssoooooshddddhyso+sddhyhdh.'
..+oooossssssssssssooooo+++//:oshdmmddhys+o++o+ohddhyhdh.'
'.-/+++ooooosssssssssssooooo/oNNdyyyyyyhhoysyhyhdddddddh..'
 '.---://+++oooooossssssssso/yMNmdddmmmmmhdddmmmmdhhso+/-..'
  .:hs+/:--::/++++oooooooooo/yMMMNNNNNNNNmmmdhys+/:---::/:.'
  '/MMMNmdhs+/:--://+++++ooo+/NMMMMNNNmdhso/::--:://+++oo+-.
  '/MMMMMMMMMNmdyo+::--://+++:yNNmdyo+/:--:://++oooooooooo:.
  '+MMMMMMMMMMMMMMMNmhyo/::--::+/:---://++oooooossssssssso:.
 '.+MMMMMMMMMMMMMMMMMMMMMNmhy+--//+++ooooosssssssssssooooo-.
 '.+mNNMMMMMMMMMMMMMMMMMMMMMMMh:ooooossssssssssssooooo+++/..
 '.--:+oydmNMMMMMMMMMMMMMMMMMMM:ossssssssssoooooo++++/::-..
'.://:-----/+shdmNMMMMMMMMMMMMM:ooooooooooo++++//:----:/-.'
..+++++++//:----:/+shmNNMMMMMMm:oooo+++++//:----:/+osyhh-.
.-oooooooo++++//:-----:/oydmNNo/+++//:-----/+oyyhddddddd-.
.-ossssssoooooooo+++//:-----:/-:-----:+oyhddhysoodddyhdd:.
..+ooossssssssssoooooo+++//::--:/oyhdmdhhso+o++ooddhyydd:.
'.-+oooooossssssssssssoooooo+/hNNdhyyssyysooyyyhhmddhddd:.'
 '.-//++++oooooosssssssssssoo/NMdhhhdddmmhhhdddmmdddhys+:..'
  ''...--://++++ooooooossssso/NMMNNNNNNNNmmmmmdhyo/:-...'''
       '''...--://+++++oooooo:dMMMMNNNNNmdhs+/--..'''
             ''''..--://+++++/oMMMNmdys+:-..''''
                   ''''..--://:syo/--..''''
                         ''''.....'''

              _ _ _     _____ _
             | (_) |   / ____| |
             | |_| |__| (___ | |_ ___  _ __ __ _  __ _  ___
             | | | '_ \\___ \| __/ _ \| '__/ _' |/ _' |/ _ \
             | | | |_) |___) | || (_) | | | (_| | (_| |  __/
             |_|_|_.__/_____/ \__\___/|_|  \__,_|\__, |\___|
                                                  __/ |
                                                 |___/

    server:      %[1]s
    admin token: %[2]s

    semver:      %[3]s
    osarch:      %[4]s
    branch:      %[5]s
    commit:      %[6]s
    formed:      %[7]s

    starting...

`
