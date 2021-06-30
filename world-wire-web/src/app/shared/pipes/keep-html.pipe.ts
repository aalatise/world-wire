import { Pipe, PipeTransform } from '@angular/core';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';

@Pipe({
    name: 'keepHtml', pure: false
})

export class EscapeHtmlPipe implements PipeTransform {

    constructor(private sanitizer: DomSanitizer) {
    }

    transform(content): SafeHtml {
        return this.sanitizer.bypassSecurityTrustHtml(content);
    }
}
