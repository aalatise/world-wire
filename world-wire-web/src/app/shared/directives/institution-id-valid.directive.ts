import { Directive, Input } from '@angular/core';
import { AsyncValidator, AbstractControl, ValidationErrors, NG_ASYNC_VALIDATORS } from '@angular/forms';
import { Observable } from 'rxjs';
import { PortalApiService } from '../services/portal-api.service';

/**
 * Checks against firebase if Institution slug already exists
 *
 * @export
 * @class InstitutionIdValidator
 * @implements {AsyncValidator}
 */
@Directive({
    selector: '[appInstitutionIdValid]',
    providers: [{
        provide: NG_ASYNC_VALIDATORS,
        useExisting: InstitutionIdValidator,
        multi: true
    }]
})
export class InstitutionIdValidator implements AsyncValidator {

    @Input('institutionSlug') institutionSlug: string;

    @Input('institutionId') institutionId: string;

    private timeout;

    constructor(
        private portalApiService: PortalApiService,
    ) { }

    validate(
        ctrl: AbstractControl
    ): Promise<ValidationErrors | null> | Observable<ValidationErrors | null> {

        // prevents submission multiple async requests
        clearTimeout(this.timeout);

        return new Promise((resolve) => {
          // delay validation to allow request to come back
          this.timeout = setTimeout(() => {
            this.portalApiService.getAllInstitution().subscribe((institutions) => {
              let slug               = null;
              const foundInstitution = institutions.some((institution) => {
                if (institution.info.institutionId === this.institutionId) {
                  return true;
                }
              });
              if (!foundInstitution) {
                slug = { slugExists: true };
              }
              resolve(slug);
            });
          }, 400);
        });
    }
}
