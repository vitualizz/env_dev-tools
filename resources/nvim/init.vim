" Directorio de plugins
call plug#begin('~/.local/share/nvim/plugged')

Plug 'tpope/vim-rails' " Manage of Rails
Plug 'tpope/vim-fugitive' " Manage Git
Plug 'scrooloose/nerdtree' " Show Tree of files
Plug 'vim-airline/vim-airline' " Bar of status
Plug 'vim-airline/vim-airline-themes' " Themes for bar of status
Plug 'airblade/vim-gitgutter' " Status of changes Git
Plug 'wakatime/vim-wakatime' " Wakatime
Plug 'posva/vim-vue' " Manage Vue
Plug 'mattn/emmet-vim' " Emmet

call plug#end()


colorscheme Tomorrow-Night-Eighties " Color of code
let g:airline_theme="tomorrow" " Set Airline Theme

set expandtab
set tabstop=2
set softtabstop=2
set shiftwidth=2

let g:user_emmet_leader_key='<C-X>'
