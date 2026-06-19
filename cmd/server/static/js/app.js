function copyAffiliateURL() {
    var input = document.getElementById('affiliate-url');
    if (!input) return;
    input.select();
    input.setSelectionRange(0, 99999);
    try {
        navigator.clipboard.writeText(input.value).then(function() {
            var btn = document.querySelector('.btn-copy');
            btn.textContent = '已复制';
            setTimeout(function() { btn.textContent = '一键复制'; }, 2000);
        }).catch(function() {
            document.execCommand('copy');
            var btn = document.querySelector('.btn-copy');
            btn.textContent = '已复制';
            setTimeout(function() { btn.textContent = '一键复制'; }, 2000);
        });
    } catch(e) {
        document.execCommand('copy');
    }
}
