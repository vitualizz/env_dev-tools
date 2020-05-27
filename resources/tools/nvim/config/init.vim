" Directorio de plugins
call plug#begin('~/.local/share/nvim/plugged')

" Languages && Frameworks
Plug 'pangloss/vim-javascript' " Javascript
Plug 'tpope/vim-rails' " Rails
Plug 'tpope/vim-fugitive' " Git
Plug 'posva/vim-vue' " Vue
Plug 'digitaltoad/vim-pug' " Pug
Plug 'tpope/vim-haml' " Haml
Plug 'kchmck/vim-coffee-script' " Coffee Script

" Plugins
Plug 'scrooloose/nerdtree' " Show Tree of files
Plug 'vim-airline/vim-airline' " Bar of status
Plug 'vim-airline/vim-airline-themes' " Themes for bar of status
Plug 'airblade/vim-gitgutter' " Status of changes Git
Plug 'wakatime/vim-wakatime' " Wakatime
Plug 'mattn/emmet-vim' " Emmet
Plug 'dense-analysis/ale' " Eslint JS

" ColorScheme
Plug 'rakr/vim-one'
Plug 'joshdick/onedark.vim'

call plug#end()

" Vim
" colorscheme Tomorrow-Night-Eighties " Color of code
" let g:airline_theme="tomorrow" " Set Airline Theme
colorscheme one
set background=dark

set expandtab
set tabstop=2
set softtabstop=2
set shiftwidth=2

set foldcolumn=2 " Margin left

set writebackup
set nobackup
" Vim End

" Airline
let g:airline_theme='one'
let g:airline_powerline_fonts = 1
" Airline End

" Eslint
let g:ale_sign_error = '❌'
let g:ale_sign_warning = '⚠️'
let g:ale_fix_on_save = 1
" Eslint End

" Emmet
let g:user_emmet_leader_key='<C-X>'
" Emmet End

" Vue
let g:vue_pre_processors = ['pug', 'sass']
" Vue End
